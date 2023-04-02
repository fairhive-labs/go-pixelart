# go-pixelart
[![Test & Heroku Deployment](https://github.com/fairhive-labs/go-pixelart/actions/workflows/heroku.yml/badge.svg)](https://github.com/fairhive-labs/go-pixelart/actions/workflows/heroku.yml)
[![Test & Docker Build+Push](https://github.com/fairhive-labs/go-pixelart/actions/workflows/docker.yml/badge.svg)](https://github.com/fairhive-labs/go-pixelart/actions/workflows/docker.yml)

### Transform picture into pixel art

:medal_military: Awesome tiny project getting fourth place :four: at [GoHack Hackathon 2022](https://twitter.com/golang/status/1545116451429818368?s=20&t=cohum9Kbf0n_kyHa8CiwWw) (1011 participants)

## HTML
> curl -vvv -X POST https://fairhive-pixelart.herokuapp.com/pixelize -F "file=@pictures/wfvr.png"  -F "edge=short" -F "slices=100" -F "filter=ega" -H "Content-Type: multipart/form-data" --output pixelart.html

## JSON
> curl -s -X POST "https://fairhive-pixelart.herokuapp.com/pixelize?mime=json" -F "file=@pictures/wfvr.png"  -F "edge=short" -F "slices=100" -F "filter=ega" -H "Content-Type: multipart/form-data" | jq

## Docker 

### Local Build & Run

`docker build -t fairhivelabs/pixelart . && docker run -it --rm -p 8080:8080 fairhivelabs/pixelart`

### Run pulling GitHub Container Registry image (ghcr.io)

`docker run -it --rm -p 8080:8080 ghcr.io/fairhive-labs/pixelart`

### Open pixelart application :rocket:

> visit `http://localhost:8080` in your web browser and pixelize every stuff ;)
