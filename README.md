# go-pixelart
[![Test & Heroku Deployment](https://github.com/fairhive-labs/go-pixelart/actions/workflows/test_heroku_deploy.yml/badge.svg)](https://github.com/fairhive-labs/go-pixelart/actions/workflows/test_heroku_deploy.yml)
[![Test & Docker Build/Push](https://github.com/fairhive-labs/go-pixelart/actions/workflows/test_docker_build_push.yml/badge.svg)](https://github.com/fairhive-labs/go-pixelart/actions/workflows/test_docker_build_push.yml)

### Transform picture into pixel art

:medal_military: Awesome tiny project getting fourth place :four: at [GoHack Hackathon 2022](https://twitter.com/golang/status/1545116451429818368?s=20&t=cohum9Kbf0n_kyHa8CiwWw) (1011 participants)

## HTML
> curl -vvv -X POST https://fairhive-pixelart.herokuapp.com/pixelize -F "file=@pictures/wfvr.png"  -F "edge=short" -F "slices=100" -F "filter=ega" -H "Content-Type: multipart/form-data" --output pixelart.html

## JSON
> curl -s -X POST "https://fairhive-pixelart.herokuapp.com/pixelize?mime=json" -F "file=@pictures/wfvr.png"  -F "edge=short" -F "slices=100" -F "filter=ega" -H "Content-Type: multipart/form-data" | jq

## Docker 

### Build
`docker build -t fairhivelabs/pixelart . && docker push fairhivelabs/pixelart`

#### Run
`docker run -it --rm -p 8080:8080 fairhivelabs/pixelart`
