# go-pixelart
transform picture into pixel art

## HTML
> curl -vvv -X POST http://localhost:8080/pixelize -F "file=@pictures/wfvr.png"  -F "edge=short" -F "slices=100" -F "filter=ega" -H "Content-Type: multipart/form-data" --output pixelart.html

## HTML
> curl -s -X POST "http://localhost:8080/pixelize?mime=json" -F "file=@pictures/wfvr.png"  -F "edge=short" -F "slices=100" -F "filter=ega" -H "Content-Type: multipart/form-data" | jq