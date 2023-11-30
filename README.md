# warhorn-stats
Uses warhorn.net's GraphQL API to generate statistics about which games are played the most.

Usage:
```command
go run main.go -token "${WARHORN_TOKEN}" | tee results.csv
```

To generate a token, create a warhorn account and follow the instructions here:
https://warhorn.net/developers/docs/products/graphql-api/overview

## Most popular game systems

Shows the number of sessions by game system. Only includes sessions with at least one GM and one player.
```command
$ awk -F, '($1 == "SESSION" && $4>0 && $5>0){s[$6]+=1} END{for(x in s) printf "%8d %s\n", s[x], x}' results.txt | sort -rn | head -n50
   93118 Dungeons & Dragons 5th Edition
   88554 Pathfinder Roleplaying Game (1st Edition)
   34354 Pathfinder Roleplaying Game (2nd Edition)
   18405 Starfinder Roleplaying Game
    4149
    2386 Board Game
    1943 Pathfinder Adventure Card Game
     791 Call of Cthulhu
     758 Dungeons & Dragons 4th Edition
     583 Custom/Freeform
     566 Savage Worlds Adventure Edition
     431 Card Game
     398 Shadowrun 4th Edition
     388 Role-playing Game
     318 Dungeon Crawl Classics RPG
     281 Torg: Eternity
     260 Miniatures Game
     244 Other
     207 Shadowrun Fifth Edition
     201 Star Wars: Edge of the Empire
     151 Shadowrun Sixth World
     149 Homebrew System
     144 Fate Core
     137 Indie RPG
     132 Pathfinder for Savage Worlds
     121 Workshop/Panel
     113 LARP
     113 Call of Cthulhu 7th Edition
     110 Advanced Dungeons & Dragons
     107 Marvel Super Heroes
     103 Dread
     100 Star Wars d6
     100 Numenera
      98 D20 System
      90 Mongoose Traveller 2nd Edition
      89 Star Wars: Age of Rebellion
      88 Shattered Empires RPG
      86 World of Darkness
      85 D&D Basic Set
      81 Social/Party Game
      79 Vampire: The Masquerade
      77 Fiasco
      74 Video Game
      70 Monster of the Week
      70 GURPS
      70 Blades in the Dark
      69 13th Age
      68 Cyberpunk Red
      64 Powered by the Apocalypse
      62 Delta Green
```

## Most active events

Confusingly, warhorn uses the term "event" to refer to a group of sessions, even though they are usually grouped by a location such as a city or a FLGS.
```command
$ awk -F, '($1 == "EVENTSUMMARY"){printf "%s,%s,%s,%s\n", $4, $5, $6, $2}' results.txt | sort -rn | head -n20 | column -t -s,
5011 sessions  149 GMs  1018 players  boston-pfs
4652 sessions  119 GMs  888 players   dragons-lair-pfs
4201 sessions  154 GMs  1108 players  indy-dnd
3502 sessions  128 GMs  1899 players  dragons-lair-austin-dnd-al
3380 sessions  90 GMs   771 players   evddal
3287 sessions  262 GMs  1321 players  sacramento-dnd-al
3132 sessions  105 GMs  662 players   ddalsa
2985 sessions  102 GMs  789 players   central-indiana-pfs
2724 sessions  100 GMs  838 players   fantasy-grounds-gaming-lounge
2705 sessions  106 GMs  862 players   sacramento-pfs
2664 sessions  154 GMs  1242 players  atomic-empire
2603 sessions  82 GMs   469 players   upstate-ny-pfs
2575 sessions  70 GMs   336 players   manitoba-pfs
2513 sessions  90 GMs   646 players   pfsnl
2438 sessions  95 GMs   434 players   dfw-pfs
2405 sessions  107 GMs  630 players   detroit-pfs
2340 sessions  78 GMs   427 players   knoxville-pfs
2242 sessions  48 GMs   371 players   ballarat-dnd
2238 sessions  55 GMs   820 players   guild-house-adventurers-league
2234 sessions  72 GMs   373 players   edmonton-pfs
```
For details on an event, go to https://warhorn.net/events/SLUG where slug is the entry in the last column.

## TODO

* Try to find a more robust way to list all events
* Visualize the top games over time (data goes back to 2013)

## See also

* Another cool thing somebody did with the API: https://github.com/michael-tracey/the-call
