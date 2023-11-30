package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

func warhornGraphqlQuery(client *retryablehttp.Client, token, query string, target interface{}) {
	// prepare request
	requestBody := "{\"query\": \"" + query + "\"}"
	req, err := retryablehttp.NewRequest(
		http.MethodPost,
		"https://warhorn.net/graphql",
		bytes.NewBuffer([]byte(requestBody)),
	)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("Content-Type", "application/json")

	// send request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err == nil {
			fmt.Print(string(body))
		}
		log.Fatalf("Response %d", resp.StatusCode)
	}

	// parse response
	if target == nil {
		target = new(struct{})
	}
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		log.Fatal(err)
	}
}

// NOTE: I couldn't find a way to list the events directly through warhorn's
// GraphQL API. So here I resort to listing all scenarios and creating the
// set of events which own those scenarios. It's possible I'm missing some
// (many?) events this way.
func getEvents(client *retryablehttp.Client, token string, handler func(string)) {
	eventsMap := make(map[string]bool)
	var cursor string
	i := 0
	for {
		cursorClause := ""
		if cursor != "" {
			cursorClause = fmt.Sprintf("(after: \\\"%s\\\")", cursor)
		}
		query := fmt.Sprintf("{globalScenarios%s{nodes{event{slug}} pageInfo{endCursor}}}", cursorClause)

		var result struct {
			Data struct {
				GlobalScenarios struct {
					Nodes []struct {
						Event struct {
							Slug string
						}
					}
					PageInfo struct {
						EndCursor string
					}
				}
			}
		}
		warhornGraphqlQuery(client, token, query, &result)

		if len(result.Data.GlobalScenarios.Nodes) == 0 {
			break
		}

		for _, node := range result.Data.GlobalScenarios.Nodes {
			slug := node.Event.Slug
			if slug != "" {
				if _, ok := eventsMap[slug]; !ok {
					// found a new event
					eventsMap[slug] = true
					handler(slug)
				}
			}
		}
		cursor = result.Data.GlobalScenarios.PageInfo.EndCursor

		i++
	}

	log.Printf("Found %d unique events after %d queries ...", len(eventsMap), i)
}

func getSessions(client *retryablehttp.Client, token, event string) {
	var cursor string
	queryCount := 0
	sessionCount := 0

	gmSet := make(map[string]bool)
	playerSet := make(map[string]bool)

	for {
		cursorClause := ""
		if cursor != "" {
			cursorClause = fmt.Sprintf(", after: \\\"%s\\\"", cursor)
		}
		query := fmt.Sprintf("{eventSessions(events: [\\\"%s\\\"]%s){nodes{startsAt scenario{gameSystem{name}} gmSignups{user{id}} gmWaitlistEntries{user{id}} playerSignups{user{id}} playerWaitlistEntries{user{id}}} pageInfo{endCursor}}}", event, cursorClause)

		type Signup struct {
			User struct {
				Id string
			}
		}
		var result struct {
			Data struct {
				EventSessions struct {
					Nodes []struct {
						StartsAt string
						Scenario struct {
							GameSystem struct {
								Name string
							}
						}
						GmSignups             []Signup
						GmWaitlistEntries     []Signup
						PlayerSignups         []Signup
						PlayerWaitlistEntries []Signup
					}
					PageInfo struct {
						EndCursor string
					}
				}
			}
		}

		warhornGraphqlQuery(client, token, query, &result)
		queryCount++

		for _, node := range result.Data.EventSessions.Nodes {
			for _, signup := range node.GmSignups {
				gmSet[signup.User.Id] = true
			}
			for _, signup := range node.GmWaitlistEntries {
				gmSet[signup.User.Id] = true
			}
			for _, signup := range node.PlayerSignups {
				playerSet[signup.User.Id] = true
			}
			for _, signup := range node.PlayerWaitlistEntries {
				playerSet[signup.User.Id] = true
			}
			fmt.Printf("SESSION,%s,%s,%d,%d,%s\n", event, node.StartsAt, len(node.GmSignups), len(node.PlayerSignups), node.Scenario.GameSystem.Name)
		}

		count := len(result.Data.EventSessions.Nodes)
		sessionCount += count
		if count == 0 {
			break
		} else if count < 100 {
			// HACK since we know the max entries per page is 100 we can skip the next query and quit early
			break
		}

		cursor = result.Data.EventSessions.PageInfo.EndCursor
	}
	fmt.Printf("EVENTSUMMARY,%s,%d queries,%d sessions,%d GMs,%d players\n", event, queryCount, sessionCount, len(gmSet), len(playerSet))
}

func main() {
	flagToken := flag.String("token", "", "auth token for warhorn graphql")
	flag.Parse()
	if *flagToken == "" {
		log.Fatal("-token is required")
	}

	client := retryablehttp.NewClient()
	client.Logger = nil
	client.RetryWaitMax = time.Hour
	client.RetryMax = 12
	client.ResponseLogHook = func(_logger retryablehttp.Logger, resp *http.Response) {
		if resp.StatusCode == 200 {
			return
		} else if resp.StatusCode == 429 {
			log.Printf("Received %s", resp.Status)
		} else {
			for key := range resp.Header {
				log.Printf("%s: %s", key, resp.Header.Get(key))
			}
			body, err := io.ReadAll(resp.Body)
			if err == nil {
				log.Print("Body:")
				fmt.Println(string(body))
			}
			log.Fatalf("Response %d", resp.StatusCode)
		}
	}

	getEvents(client, *flagToken, func(eventSlug string) {
		getSessions(client, *flagToken, eventSlug)
	})
}
