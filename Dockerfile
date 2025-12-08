# --------------------------------------------------------
# [Stage 1] Builder: 소스 코드를 빌드하는 단계
# --------------------------------------------------------
# Go 최신 버전 사용 (1.25가 아직 안 나왔으면 최신 stable 버전인 1.23 or 1.24 사용)
FROM golang:1.24-alpine AS builder

# 필수 패키지 설치 (git 등)
RUN apk add --no-cache git

# 작업 디렉토리 설정
WORKDIR /app

# 의존성 파일 복사 및 다운로드 (캐시 활용을 위해 소스보다 먼저)
COPY go.mod go.sum ./
RUN go mod download

# 소스 코드 전체 복사
COPY . .

# 빌드 실행
# - CGO_ENABLED=0: C 라이브러리 의존성 제거 (Static Linking)
# - GOOS=linux: 리눅스용 빌드
# - -ldflags="-s -w": 디버그 정보 제거하여 용량 축소
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o outbox-service cmd/worker/main.go

# --------------------------------------------------------
# [Stage 2] Runner: 실제 실행될 가벼운 이미지
# --------------------------------------------------------
FROM alpine:latest

# SSL 인증서 설치 (Kafka/DB/HTTPS 통신 필수) 및 타임존 설정
RUN apk --no-cache add ca-certificates tzdata

# 보안을 위해 루트가 아닌 사용자 생성
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# 작업 디렉토리
WORKDIR /app

# Builder 단계에서 만든 실행 파일만 쏙 빼오기
COPY --from=builder /app/outbox-service .

# 실행 명령
CMD ["./outbox-service"]