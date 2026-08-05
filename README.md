# go-pixelart

[![Test & Heroku Deployment](https://github.com/fairhive-labs/go-pixelart/actions/workflows/heroku.yml/badge.svg)](https://github.com/fairhive-labs/go-pixelart/actions/workflows/heroku.yml)
[![Test & Docker Build+Push](https://github.com/fairhive-labs/go-pixelart/actions/workflows/docker.yml/badge.svg)](https://github.com/fairhive-labs/go-pixelart/actions/workflows/docker.yml)

### Transform picture into pixel art

:medal_military: Awesome tiny project getting fourth place :four: at [GoHack Hackathon 2022](https://twitter.com/golang/status/1545116451429818368?s=20&t=cohum9Kbf0n_kyHa8CiwWw) (1011 participants)

The picture is cut into square blocks - the *slices* are counted along the short or the long edge - every block is replaced by the average color of the pixels it covers, and that color is mapped to a retro palette (CGA, EGA, VGA...) or to a grayscale/contrast filter.

Two ways to use it: an **HTTP API** (with a small web page) and a **CLI**.

---

## Quick start

### Run the image from the GitHub Container Registry

```sh
docker run -it --rm -p 8080:8080 ghcr.io/fairhive-labs/pixelart
```

### Build and run it locally

```sh
docker build -t fairhivelabs/pixelart . && docker run -it --rm -p 8080:8080 fairhivelabs/pixelart
# or
docker compose up --build
# or, without docker
make run
```

Then visit [http://localhost:8080](http://localhost:8080) :rocket: and pixelize your pictures ;)

---

## HTTP API

| Method | Path           | Description                                       |
| ------ | -------------- | ------------------------------------------------- |
| `GET`  | `/`            | web page to upload and pixelize a picture         |
| `POST` | `/pixelize`    | transform a picture, returns HTML or JSON         |
| `GET`  | `/health`      | liveness probe, returns `ok`                      |
| `GET`  | `/favicon.ico` | favicon, cached and revalidated with an `ETag`    |

### `POST /pixelize`

`multipart/form-data`, every field is required:

| Field    | Type    | Values                                                                                             | Description                                          |
| -------- | ------- | -------------------------------------------------------------------------------------------------- | ---------------------------------------------------- |
| `file`   | file    | PNG or JPEG                                                                                          | the picture to transform                             |
| `slices` | integer | `1` to `1000`, and not more than the smallest side of the picture                                    | number of blocks cut along the selected edge         |
| `edge`   | string  | `short`, `long`                                                                                      | edge the slices are counted on                       |
| `filter` | string  | `cga2`, `cga4`, `cga16`, `ega`, `vga`, `identity`, `dark-contrast`, `dark-gray`, `gray`, `invert`, `light-gray`, `xray` | color palette applied to each block |

| Query    | Values           | Description                                            |
| -------- | ---------------- | ------------------------------------------------------ |
| `mime`   | `html` (default), `json` | response format                                |

The pixel art is returned in the same format as the picture that was uploaded (a PNG in, a PNG out).

#### HTML response

Returns `201 Created` and a page displaying the pixel art, which can be rotated and resized:

```sh
curl -X POST "http://localhost:8080/pixelize" \
  -F "file=@pictures/wfvr.png" \
  -F "edge=short" \
  -F "slices=100" \
  -F "filter=ega" \
  --output pixelart.html
```

#### JSON response

Returns `201 Created` and the picture, base64 encoded:

```sh
curl -s -X POST "http://localhost:8080/pixelize?mime=json" \
  -F "file=@pictures/wfvr.png" \
  -F "edge=short" \
  -F "slices=100" \
  -F "filter=ega" | jq -r .data | base64 -d > pixelart.png
```

```json
{
  "data": "iVBORw0KGgoAAAANSUhEUgAAB+QAAAp4CAIAAADg...",
  "encoding": "base64",
  "filter": "ega",
  "length": 36192
}
```

#### Status codes

| Code                      | When                                                                                     |
| ------------------------- | ---------------------------------------------------------------------------------------- |
| `201 Created`             | the pixel art was produced                                                                 |
| `400 Bad Request`         | a field is missing or invalid, or `slices` is larger than the picture                      |
| `413 Content Too Large`   | the upload is over `MAX_UPLOAD_BYTES`, or the picture declares more than `MAX_PIXELS`      |
| `415 Unsupported Media Type` | the file is not a PNG or a JPEG, or cannot be decoded                                   |
| `503 Service Unavailable` | `MAX_CONCURRENT_JOBS` transformations are already running, retry later (`Retry-After`)     |

The API is public and CORS is open to any origin (`POST`, `OPTIONS`), without credentials.

---

## Filters

Applied to the average color of each block. `pictures/wfvr.png` is the source of every sample below.

| Filter          | Description                                        | Sample                                                    |
| --------------- | -------------------------------------------------- | --------------------------------------------------------- |
| `cga2`          | 2 colors CGA palette, black and white              | [wfvr_cga2.png](pictures/wfvr_cga2.png)                     |
| `cga4`          | 4 colors CGA palette                               | [wfvr_cga4.png](pictures/wfvr_cga4.png)                     |
| `cga16`         | 16 colors CGA palette                              | [wfvr_cga16.png](pictures/wfvr_cga16.png)                   |
| `ega`           | 64 colors EGA palette                              | [wfvr_cga64.png](pictures/wfvr_cga64.png)                   |
| `vga`           | 18 bits VGA color, 6 bits per channel              | -                                                           |
| `identity`      | true color, no palette                             | [wfvr_pixel80_shortedge_truecolor.png](pictures/wfvr_pixel80_shortedge_truecolor.png) |
| `gray`          | grayscale, average of the channels                 | -                                                           |
| `dark-gray`     | grayscale using the darkest channel                | [wfvr_dark_gray.png](pictures/wfvr_dark_gray.png)           |
| `light-gray`    | grayscale using the brightest channel              | -                                                           |
| `dark-contrast` | darkens the two brightest channels                 | [wfvr_dark_contrast.png](pictures/wfvr_dark_contrast.png)   |
| `invert`        | inverted colors                                    | -                                                           |
| `xray`          | inverted, then brightest channel                   | -                                                           |

Transparent pixels stay transparent with the palette filters.

---

## CLI

```sh
make cli          # builds ./bin/pixelart
./bin/pixelart -slices 80 -edge long -filter cga16 pictures/wfvr.png
# or without building
go run ./cmd -slices 80 -edge long -filter cga16 pictures/wfvr.png
```

| Flag       | Default                          | Description                                          |
| ---------- | -------------------------------- | ---------------------------------------------------- |
| `-slices`  | `100`                            | number of slices cut along the selected edge         |
| `-edge`    | `short`                          | `short` or `long`                                    |
| `-filter`  | `cga4`                           | any filter of the table above                        |
| `-out`     | `<picture>_<timestamp>.<ext>`    | output file                                          |
| `-quality` | `75`                             | JPEG output quality, from 1 to 100                   |

```
👉 Source file "pictures/wfvr.png" opened
🤖 Image DECODED - Format is "png"
🖼  Original Dimension = [ 2010 x 2679 ]
👾 Processing Transformation...
✅ Transformation is over
💾 Pixel Art saved in file "pictures/wfvr_20260804-233956.png"
```

Exit codes: `0` on success, `2` on a usage error, `1` when the transformation fails.

---

## Configuration

Every setting is an environment variable read at startup:

| Variable              | Default            | Description                                                             |
| --------------------- | ------------------ | ----------------------------------------------------------------------- |
| `PORT`                | `8080`             | port the server listens on                                              |
| `MAX_UPLOAD_BYTES`    | `10485760` (10 MiB) | largest accepted request body                                          |
| `MAX_PIXELS`          | `40000000` (40 MP) | largest accepted picture, checked before decoding                       |
| `MAX_CONCURRENT_JOBS` | number of CPUs     | transformations running at the same time, further requests get a `503`  |
| `GIN_MODE`            | `release`          | set to `debug` for gin's verbose mode                                   |

The dimensions are read from the picture header **before** it is decoded, so an oversized picture is rejected without allocating memory for it.

---

## Development

```sh
make build   # builds the API into ./bin/api
make run     # builds and runs it on :8080
make cli     # builds the CLI into ./bin/pixelart
make test    # go test ./...
make race    # go test -race ./...
make bench   # filter benchmarks
make vet     # go vet ./...
```

Go 1.25 or later is required (see `go.mod`).

Layout:

| Path                  | Content                                                     |
| --------------------- | ----------------------------------------------------------- |
| `api/`                | HTTP server, embedded templates and assets                  |
| `cmd/`                | command line tool                                           |
| `internal/filter/`    | pixel, basic and convolution filters, CGA/EGA/VGA palettes  |
| `internal/colorutils/`| color helpers                                               |
| `pictures/`           | sample pictures                                             |

The filters are parallelized over the available CPUs; `internal/filter/equivalence_test.go` keeps the original implementations as an oracle and checks the output pixel per pixel, so an optimization that changes the result fails the build.

---

## Deployment

### Docker

The image builds a static binary and runs it as an unprivileged user on a pinned alpine, with a health check on `/health`:

```sh
docker build -t fairhivelabs/pixelart .
docker run -it --rm -p 8080:8080 -e MAX_UPLOAD_BYTES=20971520 fairhivelabs/pixelart
```

### GitHub Container Registry

Every push on `main` publishes `ghcr.io/fairhive-labs/pixelart` (see `.github/workflows/docker.yml`).

### Heroku

`.github/workflows/heroku.yml` deploys the Docker image on every push on `main`. The deployment is skipped when the `HEROKU_API_KEY` secret is missing. Set the `HEROKU_APP_NAME` and `HEROKU_EMAIL` repository variables to target another app - the health check URL follows the app name.

---

## License

[MIT](LICENSE)
