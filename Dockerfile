# --------------------------------------------------------
# [Stage 1] Builder
# --------------------------------------------------------
# 1. 버전 올림: go.mod가 1.25를 요구하므로 이미지도 1.25로 변경
FROM golang:1.25-alpine AS builder

# 필수 패키지 설치
RUN apk add --no-cache git

# 작업 디렉토리 설정
WORKDIR /app

# 소스 코드 복사
COPY . .

# 2. 삭제함: `myapp` 빌드 라인은 불필요하며 에러 유발 가능성 있음
# RUN go build -mod=vendor -o myapp main.go  <-- 삭제!

# 3. 실제 빌드 실행 (outbox-service)
# vendor 모드 사용 시 -mod=vendor 옵션 추가 필요
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -ldflags="-s -w" -o outbox-service cmd/worker/main.go

# --------------------------------------------------------
# [Stage 2] Runner
# --------------------------------------------------------
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

WORKDIR /app

# Builder에서 생성한 바이너리 복사
COPY --from=builder /app/outbox-service .

CMD ["./outbox-service"]