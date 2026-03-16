FROM golang:1.24-alpine AS builder
RUN apk add --no-cache gcc musl-dev sqlite-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o /bin/refactorkit .

FROM python:3.12-alpine
RUN apk add --no-cache sqlite-libs && \
    pip install --no-cache-dir nltk && \
    python3 -c "import nltk; nltk.download('punkt_tab', quiet=True); nltk.download('stopwords', quiet=True)"
COPY --from=builder /bin/refactorkit /bin/refactorkit
ENTRYPOINT ["/bin/refactorkit"]
