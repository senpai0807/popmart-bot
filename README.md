# Popmart CLI Bot
- I got bored and wanted to create a CLI bot using Golang, I made it for myself and I've secured like 10 or so BIEs and I made this overnight. I was running 500 accounts/tasks for the trolls, but you can scale higher.

## Features
- Login Session Storage (stores session in sessions.json)
- Adyen encryption for card payment (includes risk data generation)
- Supports both card and paypal payments
- Discord webhooks for paypal checkout links and success
- TD Solver is integrated locally and bundled within the exe upon building
- IMAP Integration (I have it integrated to fetch the codes, but i didn't feel like integrating a generator)

## How To Use
```
1. git clone https://github.com/senpai0807/popmart-bot.git
2. cd popmart-bot
3. go run ./src/main.go
```
## Updates Needed
- I've placed comments, but it should run as is, you just won't have the auto update logic. To use it, you'll need to use a CDN like Digital Oceans and upload a version.json file and exe and paste the corresponding urls for each in update.go
- You'll also need to bundle the updater.go into an exe and upload it into the cdn and copy and paste the url in helpers.go in the downloadUpdater func
