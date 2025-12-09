job "outbox-worker" {
  datacenters = ["dc1"]
  type        = "service"

  group "worker-group" {
    count = 1

    task "go-worker" {
      driver = "docker"

      config {
        image = "suehunpark/outbox-service:latest" # 도커 허브 이미지
        force_pull = true
      }

      env {
        # 1. Kafka 주소
        KAFKA_BROKERS = "10.0.4.2:9092,10.0.4.5:9092"

        # 2. 계좌 DB (5432 포트)
        # postgres://아이디:비번@10.0.4.2:5432/DB명
        ACCOUNT_DB_DSN = "postgres://hts:hts@10.0.4.2:5432/hts_account?sslmode=disable"

        # 3. 주문 DB (5434 포트 - 아웃박스 폴링 대상)
        ORDER_DB_DSN   = "postgres://hts:hts@10.0.4.2:5434/hts_order?sslmode=disable"
        
        LOG_LEVEL = "WARN"
      }

      resources {
        cpu    = 400 # 0.2 core
        memory = 64  # 64 MB
      }
    }
  }
}
