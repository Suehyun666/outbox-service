job "outbox-worker" {
  datacenters = ["dc1"]
  type        = "service"

  group "worker-group" {
    count = 2 # 레플리카 2개 (고가용성)

    task "go-worker" {
      driver = "docker"

      config {
        image = "suehunpark/outbox-service:latest" # 도커 허브 이미지
        force_pull = true
      }

      # [핵심] 여기서 주입합니다!
      env {
        # 1. Kafka 주소
        KAFKA_BROKERS = "10.0.4.10:9092"

        # 2. 계좌 DB (5432 포트)
        # postgres://아이디:비번@10.0.4.2:5432/DB명
        ACCOUNT_DB_DSN = "postgres://postgres:secret1234@10.0.4.2:5432/account_db?sslmode=disable"

        # 3. 주문 DB (5434 포트 - 아웃박스 폴링 대상)
        ORDER_DB_DSN   = "postgres://postgres:secret1234@10.0.4.2:5434/order_db?sslmode=disable"
      }

      resources {
        cpu    = 200 # 0.2 core
        memory = 64  # 64 MB
      }
    }
  }
}