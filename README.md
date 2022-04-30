# go-pixelart
[![Test & Heroku Deployment](https://github.com/fairhive-labs/go-pixelart/actions/workflows/test_heroku_deploy.yml/badge.svg)](https://github.com/fairhive-labs/go-pixelart/actions/workflows/test_heroku_deploy.yml)
[![Test & Docker Build/Push](https://github.com/fairhive-labs/go-pixelart/actions/workflows/test_docker_build_push.yml/badge.svg)](https://github.com/fairhive-labs/go-pixelart/actions/workflows/test_docker_build_push.yml)

transform picture into pixel art

## HTML
> curl -vvv -X POST https://fairhive-pixelart.herokuapp.com/pixelize -F "file=@pictures/wfvr.png"  -F "edge=short" -F "slices=100" -F "filter=ega" -H "Content-Type: multipart/form-data" --output pixelart.html

## JSON
> curl -s -X POST "https://fairhive-pixelart.herokuapp.com/pixelize?mime=json" -F "file=@pictures/wfvr.png"  -F "edge=short" -F "slices=100" -F "filter=ega" -H "Content-Type: multipart/form-data" | jq

## Docker 

### Build
`docker build -t fairhivelabs/pixelart . && docker push fairhivelabs/pixelart`

#### Run
`docker run -it --rm -p 8080:8080 fairhivelabs/pixelart`
