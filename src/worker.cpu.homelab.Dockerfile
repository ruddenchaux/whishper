# Homelab build of the anysub worker. Build from the repo root:
#
#   docker build -f src/worker.cpu.homelab.Dockerfile -t ghcr.io/ruddenchaux/anysub-worker:<tag> .
#
# BASE_IMAGE supplies the Python/whisperx runtime. It defaults to the last
# known-good worker image so whisperx and numpy stay at the versions that
# work today; to rebuild the runtime from scratch, build
# whisperx-api/Dockerfile.cpu and pass it as --build-arg BASE_IMAGE=...
ARG BASE_IMAGE=ghcr.io/ruddenchaux/anysub-worker:cpu-95b7915-vad

FROM docker.io/library/node:22-alpine AS css
WORKDIR /repo
COPY package.json package-lock.json tailwind.config.js ./
RUN npm ci
COPY src/frontend ./src/frontend
RUN npx tailwindcss -i src/frontend/static/css/input.css -o src/frontend/static/css/style.css --minify

FROM docker.io/library/golang:alpine AS builder
WORKDIR /app
COPY src/go.mod src/go.sum ./
RUN go mod download && go install github.com/a-h/templ/cmd/templ@v0.2.707
COPY src/ .
RUN templ generate && GOOS=linux go build -o anysub .

FROM ${BASE_IMAGE}
# The whisperx-api Python sources are pure Python, so they can be refreshed
# without rebuilding the pip layers of the base image.
COPY whisperx-api/main.py whisperx-api/classes.py whisperx-api/transcribe.py /app/
COPY whisperx-api/backends/ /app/backends/
COPY --from=builder /app/anysub /usr/bin/anysub
COPY --from=builder /app/frontend /app/frontend
COPY --from=css /repo/src/frontend/static/css/style.css /app/frontend/static/css/style.css
COPY src/worker.entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
