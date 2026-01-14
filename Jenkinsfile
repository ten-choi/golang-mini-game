pipeline {
    agent any

    options {
        timestamps()
    }

    environment {
        // SSH
        MANAGE_HOST = "www-web@manage-server.coconefk.jp"
        TARGET_HOST = "hange-dev-game-ap-0.coconefk"

        // Git
        GIT_URL = "https://github.com/yoru-choi/golang-mini-game.git"
        BRANCH  = "dev"

        // GHCR
        GHCR_CRED_ID = "quiz-github-pat-for-ghcr"
        IMAGE_NAME   = "ghcr.io/yoru-choi/go-server-app:v1.0"

        // Docker runtime
        CONTAINER_NAME = "go-server-app"
        HOST_PORT      = "30080"
        CONTAINER_PORT = "8080"

        // ✅ .env 파일 경로 (Jenkins가 자동 배포)
        ENV_FILE = "/home/hange-dev-game-ap-0/.env"
        
        // ✅ 환경 변수 (docker-compose.yml과 일치)
        SERVER_PORT = "8080"
        GIN_MODE = "release"
        MONGO_URI = "mongodb://admin:password@10.33.255.96:27017"
        MONGO_DB = "draw_and_guess_db"
        VALKEY_ADDR = "10.33.255.96:6379"
        VALKEY_PASSWORD = "game123"
    }

    stages {
        stage('0. Checkout Source') {
            steps {
                echo "[0/2] Cloning ${BRANCH} branch..."
                checkout([$class: 'GitSCM',
                    branches: [[name: "*/${BRANCH}"]],
                    userRemoteConfigs: [[url: "${GIT_URL}", credentialsId: "${GHCR_CRED_ID}"]]
                ])
            }
        }

        stage('1. Deploy .env File (Optional)') {
            when {
                expression { env.ENV_FILE != "__NONE__" }
            }
            steps {
                echo "[1/2] Deploying .env file to target host..."
                sh """
                    set -e
                    
                    # .env 파일 내용 생성
                    cat > .env.tmp << 'ENVFILE'
SERVER_PORT=8080
GIN_MODE=release

MONGO_URI=mongodb://admin:password@10.33.255.96:27017
MONGO_DB=draw_and_guess_db

VALKEY_ADDR=10.33.255.96:6379
VALKEY_PASSWORD=game123
ENVFILE

                    # 서버로 복사
                    ssh -i ~/.ssh/id_rsa -o StrictHostKeyChecking=no ${MANAGE_HOST} \\
                      "ssh -o StrictHostKeyChecking=no ${TARGET_HOST} 'cat > ${ENV_FILE}'" < .env.tmp
                    
                    rm -f .env.tmp
                    echo "✅ .env file deployed to ${ENV_FILE}"
                """
            }
        }

        stage('2. Deploy Docker Container') {
            steps {
                echo "[2/2] Pull & Run Docker image on target host..."

                withCredentials([usernamePassword(
                    credentialsId: "${GHCR_CRED_ID}",
                    usernameVariable: 'GHCR_USER',
                    passwordVariable: 'GHCR_TOKEN'
                )]) {
                    sh """
                        set -e

                        ssh -i ~/.ssh/id_rsa -o StrictHostKeyChecking=no ${MANAGE_HOST} \\
                          "ssh -o StrictHostKeyChecking=no ${TARGET_HOST} 'bash -s'" << EOF
                            set -euo pipefail

                            IMAGE_NAME="${IMAGE_NAME}"
                            CONTAINER_NAME="${CONTAINER_NAME}"
                            HOST_PORT="${HOST_PORT}"
                            CONTAINER_PORT="${CONTAINER_PORT}"
                            ENV_FILE="${ENV_FILE}"

                            echo "🔐 Logging into GHCR..."
                            echo "${GHCR_TOKEN}" | docker login ghcr.io -u "${GHCR_USER}" --password-stdin

                            echo "📦 Pulling image: \$IMAGE_NAME"
                            docker pull "\$IMAGE_NAME"

                            echo "🧹 Removing old container (if exists): \$CONTAINER_NAME"
                            docker rm -f "\$CONTAINER_NAME" >/dev/null 2>&1 || true

                            echo "🚀 Running new container..."
                            
                            # .env 파일 사용 방식
                            if [ "\$ENV_FILE" != "__NONE__" ] && [ -f "\$ENV_FILE" ]; then
                              echo "✅ Using env-file: \$ENV_FILE"
                              docker run -d \\
                                --name "\$CONTAINER_NAME" \\
                                --restart unless-stopped \\
                                -p "\$HOST_PORT:\$CONTAINER_PORT" \\
                                --env-file "\$ENV_FILE" \\
                                "\$IMAGE_NAME"
                            else
                              # 환경 변수 직접 주입 방식
                              echo "⚠️  No env-file, using direct env vars"
                              docker run -d \\
                                --name "\$CONTAINER_NAME" \\
                                --restart unless-stopped \\
                                -p "\$HOST_PORT:\$CONTAINER_PORT" \\
                                -e SERVER_PORT="${SERVER_PORT}" \\
                                -e GIN_MODE="${GIN_MODE}" \\
                                -e MONGO_URI="${MONGO_URI}" \\
                                -e MONGO_DB="${MONGO_DB}" \\
                                -e VALKEY_ADDR="${VALKEY_ADDR}" \\
                                -e VALKEY_PASSWORD="${VALKEY_PASSWORD}" \\
                                "\$IMAGE_NAME"
                            fi

                            echo "✅ Container started!"
                            docker ps --filter "name=\$CONTAINER_NAME" --format "table {{.Names}}\\t{{.Image}}\\t{{.Status}}\\t{{.Ports}}"

                            echo ""
                            echo "📍 Service URL: http://hange-dev-game-ap-0.coconefk:\$HOST_PORT"
                            echo "📍 GraphQL: http://hange-dev-game-ap-0.coconefk:\$HOST_PORT/graphql"
                            
                            # 컨테이너 로그 확인
                            echo ""
                            echo "📋 Container logs (last 20 lines):"
                            docker logs --tail 20 "\$CONTAINER_NAME"
EOF
                    """
                }
            }
        }
    }

    post {
        success {
            echo "✅ 배포 성공! Docker 컨테이너가 정상 실행 중입니다."
            echo "📍 URL: http://hange-dev-game-ap-0.coconefk:30080"
            echo "📍 GraphQL: http://hange-dev-game-ap-0.coconefk:30080/graphql"
            echo "📍 Health: http://hange-dev-game-ap-0.coconefk:30080/health"
        }
        failure {
            echo "❌ 배포 실패. 아래 명령어로 디버깅하세요:"
            echo "  ssh hange-dev-game-ap-0.coconefk"
            echo "  docker ps -a"
            echo "  docker logs -n 200 go-server-app"
            echo "  docker inspect go-server-app"
        }
    }
}
