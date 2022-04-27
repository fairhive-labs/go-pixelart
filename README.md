# go-pixelart
transform picture into pixel art

## HTML
> curl -vvv -X POST https://fairhive-pixelart.herokuapp.com/pixelize -F "file=@pictures/wfvr.png"  -F "edge=short" -F "slices=100" -F "filter=ega" -H "Content-Type: multipart/form-data" --output pixelart.html

## JSON
> curl -s -X POST "https://fairhive-pixelart.herokuapp.com/pixelize?mime=json" -F "file=@pictures/wfvr.png"  -F "edge=short" -F "slices=100" -F "filter=ega" -H "Content-Type: multipart/form-data" | jq

## Docker 

### Build
`docker build -t fairhive-labs/pixelart . && docker push fairhive-labs/pixelart`

#### Run
`docker run -it --rm fairhive-labs/pixelart`