# 🔧 로컬 개발 환경 설정 가이드

## 📋 개요

로컬과 서버 모두 **10.33.255.96**의 Docker Compose Valkey와 MongoDB를 사용합니다.

```
┌──────────────────────────────────────────┐
│  로컬 개발 PC (Windows)                   │
│  ┌────────────────────────────────────┐  │
│  │  Go Server (localhost:8080)        │  │
│  │  React Client (localhost:5174)     │  │
│  └────────────────────────────────────┘  │
│           ↓ 연결                          │
└──────────────────────────────────────────┘
           ↓
    Internet/Network
           ↓
┌──────────────────────────────────────────┐
│  서버 (10.33.255.96)                      │
│  ┌────────────────────────────────────┐  │
│  │  Docker Compose                    │  │
│  │  • Valkey :6379                    │  │
│  │  • MongoDB :27017                  │  │
│  └────────────────────────────────────┘  │
└──────────────────────────────────────────┘
```

## 🚀 설정 단계

### 1. 서버에서 Docker Compose 시작

**서버 (10.33.255.96)에서:**
```bash
cd /home/yeol/draw-and-guess-server/deployments

# 환경 변수 설정
cp .env.example .env
nano .env

# 비밀번호 설정 (중요!)
# VALKEY_PASSWORD=your_secure_password_123
# MONGO_ROOT_PASSWORD=mongo_root_pass_456
```

**Docker 시작:**
```bash
docker-compose up -d

# 상태 확인
docker-compose ps
docker-compose logs -f
```

### 2. 로컬 Go 서버 설정

**파일: `draw-and-guess-server/.env`**

이미 업데이트되었습니다:
```env
# Server Configuration
SERVER_PORT=8080
GIN_MODE=debug

# Database Configuration (Docker Compose on 10.33.255.96)
MONGO_URI=mongodb://admin:password@10.33.255.96:27017
MONGO_DB=draw_and_guess_db

# Cache Configuration (Docker Compose on 10.33.255.96)
VALKEY_ADDR=10.33.255.96:6379
VALKEY_PASSWORD=your_secure_password_123
```

⚠️ **중요**: `VALKEY_PASSWORD`를 서버의 `deployments/.env`와 동일하게 설정하세요!

### 3. 코드 업데이트 완료

다음 파일들이 자동으로 업데이트되었습니다:
- ✅ `internal/config/config.go` - ValkeyPassword 변수 추가
- ✅ `internal/valkey/valkey.go` - 비밀번호 사용 설정
- ✅ `.env` - 새 서버 주소로 업데이트

### 4. 서버 재빌드 및 실행

```powershell
cd c:\workSpace\projects\personal\golang\golang-mini-game\draw-and-guess-server

# 재빌드
go build -o bin/server.exe ./cmd/server

# 실행
.\bin\server.exe

# 또는 백그라운드로
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd c:\workSpace\projects\personal\golang\golang-mini-game\draw-and-guess-server; .\bin\server.exe" -WindowStyle Minimized
```

### 5. 클라이언트 실행

```powershell
cd c:\workSpace\projects\personal\golang\golang-mini-game\draw-and-guess-client

# 개발 서버 시작
npm run dev
```

브라우저: http://localhost:5174

## ✅ 연결 확인

### Go 서버 로그 확인

서버 시작 시 다음과 같은 로그가 보여야 합니다:
```
Loaded configuration from .env file
Config loaded - ServerPort: 8080, ValkeyAddr: 10.33.255.96:6379, MongoDB: mongodb://admin:password@10.33.255.96:27017/draw_and_guess_db
✓ Connected to Valkey
✓ Connected to MongoDB
```

### 수동 테스트

**Valkey 연결:**
```powershell
# Windows에서 (valkey-cli 설치 필요)
valkey-cli -h 10.33.255.96 -p 6379 -a your_password ping
# 응답: PONG
```

**MongoDB 연결:**
```powershell
# mongosh 설치 필요
mongosh "mongodb://admin:password@10.33.255.96:27017/draw_and_guess_db"
# 성공 시 MongoDB 쉘 진입
```

**Health Check:**
```powershell
Invoke-RestMethod http://localhost:8080/health
# 응답: {"status":"healthy","timestamp":"..."}
```

## 🔐 보안 주의사항

### 방화벽 설정 (서버)

서버에서 포트를 열어야 합니다:
```bash
# 포트 확인
sudo firewall-cmd --list-ports

# 포트 개방
sudo firewall-cmd --add-port=6379/tcp --permanent
sudo firewall-cmd --add-port=27017/tcp --permanent
sudo firewall-cmd --reload
```

### 비밀번호 관리

1. **.env 파일을 Git에 커밋하지 마세요!**
   - 이미 `.gitignore`에 포함되어 있어야 합니다.

2. **강력한 비밀번호 사용:**
   ```bash
   # 랜덤 비밀번호 생성
   openssl rand -base64 32
   ```

3. **프로덕션 환경:**
   - 환경 변수로 비밀번호 설정
   - Secrets 관리 도구 사용 (AWS Secrets Manager, HashiCorp Vault 등)

## 🐛 문제 해결

### 1. "connection refused" 오류

**증상:** `dial tcp 10.33.255.96:6379: connect: connection refused`

**해결:**
```bash
# 서버에서
docker-compose ps  # 컨테이너 실행 상태 확인
docker-compose logs valkey  # 로그 확인

# 방화벽 확인
sudo firewall-cmd --list-ports
sudo netstat -tlnp | grep 6379
```

### 2. "authentication failed" 오류

**증상:** `ERR invalid password`

**해결:**
1. 서버 `deployments/.env`의 `VALKEY_PASSWORD` 확인
2. 로컬 `.env`의 `VALKEY_PASSWORD`와 일치하는지 확인
3. Docker 재시작:
   ```bash
   docker-compose restart valkey
   ```

### 3. MongoDB 연결 실패

**증상:** `Authentication failed`

**해결:**
```bash
# 서버에서
docker-compose exec mongodb mongosh -u root -p

# MongoDB에서
use draw_and_guess_db
db.getUsers()

# 사용자 재생성이 필요하면
db.createUser({
  user: 'admin',
  pwd: 'password',
  roles: ['readWrite', 'dbAdmin']
})
```

### 4. 네트워크 지연

**증상:** 느린 응답 시간

**해결:**
```bash
# 네트워크 레이턴시 확인
ping 10.33.255.96

# VPN/프록시 확인
# .env에서 타임아웃 조정 (valkey.go 참조)
```

## 📊 개발 워크플로우

### 일상적인 개발

1. **서버 확인** (한 번만)
   ```bash
   # 서버에서
   docker-compose ps
   # 둘 다 "Up (healthy)" 상태여야 함
   ```

2. **로컬 개발 시작**
   ```powershell
   # Terminal 1: Go 서버
   cd draw-and-guess-server
   .\bin\server.exe

   # Terminal 2: React 클라이언트
   cd draw-and-guess-client
   npm run dev
   ```

3. **코드 변경 후**
   - Go 코드: 서버 재시작 필요
   - React 코드: HMR로 자동 리로드

### 데이터 리셋

```bash
# 서버에서 - 모든 데이터 삭제
docker-compose down -v
docker-compose up -d

# 또는 - 특정 컬렉션만 삭제
docker-compose exec mongodb mongosh -u admin -p password draw_and_guess_db
> db.game_rooms.deleteMany({})
> db.users.deleteMany({})
```

## 🎯 체크리스트

로컬 개발 환경 준비:

### 서버 (10.33.255.96)
- [ ] Docker Compose 실행 중
- [ ] Valkey 헬스체크 통과
- [ ] MongoDB 헬스체크 통과
- [ ] 방화벽 포트 6379, 27017 개방
- [ ] `deployments/.env` 비밀번호 설정

### 로컬 PC
- [ ] `.env` 파일 업데이트 (IP, 비밀번호)
- [ ] Go 서버 재빌드
- [ ] Go 서버 시작 성공
- [ ] Health check 통과
- [ ] React 클라이언트 시작
- [ ] 브라우저에서 접속 확인

## 📚 관련 문서

- [MIGRATION.md](./MIGRATION.md) - K8s에서 Docker로 마이그레이션
- [deployments/README.md](./deployments/README.md) - Docker Compose 사용법
- [deployments/docker-compose.yml](./deployments/docker-compose.yml) - Docker 설정

## 💡 팁

### VS Code 통합

**launch.json** 추가:
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch Server",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/cmd/server",
      "envFile": "${workspaceFolder}/.env"
    }
  ]
}
```

### PowerShell 별칭

**프로필 추가** (`$PROFILE`):
```powershell
function Start-GameServer {
    cd c:\workSpace\projects\personal\golang\golang-mini-game\draw-and-guess-server
    .\bin\server.exe
}

function Start-GameClient {
    cd c:\workSpace\projects\personal\golang\golang-mini-game\draw-and-guess-client
    npm run dev
}

Set-Alias game-server Start-GameServer
Set-Alias game-client Start-GameClient
```

사용: `game-server`, `game-client`
