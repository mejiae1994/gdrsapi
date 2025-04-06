# gdrsAPI

A lightweight API built with mostly standard library. This API powers the game design document generator and steam rating tool found in gamedevreststop.com

## Features
- /getsteamrating endpoint scrapes and rates a video game steam page.
- /gengamedesigndoc endpoint takes some input and generates game design document content using LLM tech

## Dependencies
- Go 1.23.1
- Docker (optional, for containerization)
- Cloudflare Llava API for Img to Text
- Gemini Flash API for content evaluation
- Google service account for google sheets(not required)

## Get Started
- Need to create `.env` and put your own api keys based on `.env.example`
- run `make r` or `go run ./cmd/api` to get app running
