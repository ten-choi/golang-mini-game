# ----------------------------------------------------
# 1단계: 빌더 (Builder Stage) - 컴파일 환경
# ----------------------------------------------------
# 최신 Go 버전을 사용합니다. (Go 1.24-alpine 유지)
FROM golang:1.24-alpine AS builder

# 작업 디렉토리 설정
WORKDIR /app

# git, tzdata 등 불필요한 설치를 최소화하고, go mod download 및 build에 필요한 파일만 복사
# tzdata는 최종 alpine 이미지에 설치하므로 여기서는 제거합니다.
RUN apk add --no-cache ca-certificates

# Go 모듈 파일을 복사하고 의존성을 다운로드합니다. (캐싱 활용)
COPY go.mod go.sum ./
RUN go mod download

# 모든 소스 코드를 복사합니다.
COPY . .

# 정적 컴파일 환경 설정
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# 애플리케이션 컴파일.
# 실행 파일 이름은 'server'입니다.
RUN go build -o /server -ldflags="-s -w" -trimpath ./main.go

# ----------------------------------------------------
# 2단계: 최종 이미지 (Final Stage) - 실행 환경 (보안 및 경량화)
# ----------------------------------------------------
# 경량화와 디버깅 편의성을 위해 alpine을 사용합니다.
FROM alpine:latest

# CA 인증서와 타임존 데이터 설치 (HTTPS 통신 및 시간 처리를 위해 필수)
# tzdata를 설치하면 /usr/share/zoneinfo에 타임존 정보가 포함됩니다.
RUN apk --no-cache add ca-certificates tzdata

# 비-루트 사용자 및 그룹 생성 (보안 강화)
RUN adduser -D -g '' appuser
USER appuser

# 작업 디렉토리 설정 (비-루트 사용자 권한으로)
WORKDIR /home/appuser

# 빌더 단계에서 컴파일된 실행 파일만 복사하고 실행 권한 부여
# 'appuser'가 소유하도록 복사합니다.
COPY --from=builder --chown=appuser:appuser /server ./server

# 서버가 리스닝할 포트를 외부에 알립니다.
EXPOSE 8080

# 헬스체크 추가 (기존과 동일하게 유지)
# 참고: wget 대신 curl이 더 일반적일 수 있으나, wget도 alpine에 포함되어 사용 가능합니다.
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 컨테이너가 시작될 때 실행될 명령어입니다.
CMD ["./server"]