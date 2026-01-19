# 🎨 Draw and Guess - Frontend Client

React + TypeScript 기반 실시간 멀티플레이어 그림 맞추기 게임 클라이언트

## 📋 목차

- [기술 스택](#기술-스택)
- [프로젝트 구조](#프로젝트-구조)
- [시작하기](#시작하기)
- [환경 변수](#환경-변수)
- [개발 가이드](#개발-가이드)
- [빌드 및 배포](#빌드-및-배포)
- [주요 기능](#주요-기능)

## 🛠️ 기술 스택

- **Framework**: React 18.2
- **Language**: TypeScript 5.3
- **Build Tool**: Vite 5.0
- **Routing**: React Router 6.20
- **HTTP Client**: Axios 1.6
- **WebSocket**: Native WebSocket API
- **Styling**: CSS (inline styles)

## 📁 프로젝트 구조

```
draw-and-guess-client/
├── src/
│   ├── App.tsx                 # 메인 앱 컴포넌트 및 라우팅
│   ├── main.tsx                # 앱 진입점
│   ├── App.css                 # 전역 스타일
│   │
│   ├── components/             # 재사용 가능한 컴포넌트
│   │   └── DrawingCanvas.tsx   # 그리기 캔버스 컴포넌트
│   │
│   ├── pages/                  # 페이지 컴포넌트
│   │   ├── Home.tsx            # 홈 페이지 (입장 화면)
│   │   ├── RoomList.tsx        # 방 목록 페이지
│   │   └── GameRoom.tsx        # 게임 방 페이지 (메인 게임)
│   │
│   ├── services/               # API 및 WebSocket 서비스
│   │   ├── api.ts              # REST API 클라이언트
│   │   └── websocket.ts        # WebSocket 클라이언트
│   │
│   ├── types/                  # TypeScript 타입 정의
│   │   └── index.ts            # 공통 타입 정의
│   │
│   └── i18n/                   # 다국어 지원
│       ├── LanguageContext.tsx # 언어 컨텍스트
│       └── translations.ts     # 번역 데이터
│
├── index.html                  # HTML 템플릿
├── package.json                # 프로젝트 설정 및 의존성
├── tsconfig.json               # TypeScript 설정
├── vite.config.ts              # Vite 설정
└── README.md                   # 이 파일
```

## 🚀 시작하기

### 사전 요구사항

- Node.js 18.0 이상
- npm 9.0 이상
- 백엔드 서버가 실행 중이어야 합니다 (http://localhost:8080)

### 설치

```bash
# 의존성 설치
npm install
```

### 개발 서버 실행

```bash
# 개발 서버 시작 (기본 포트: 5173)
npm run dev

# 네트워크의 다른 기기에서 접속 가능하게 실행
npm run dev -- --host
```

브라우저에서 http://localhost:5173 접속

### 빌드

```bash
# 프로덕션 빌드
npm run build

# 빌드 결과 미리보기
npm run preview
```

## 🔧 환경 변수

프로젝트 루트에 `.env` 파일을 생성하여 환경 변수를 설정할 수 있습니다.

### .env 예시

```env
# API 서버 URL
VITE_API_BASE_URL=http://localhost:8080

# WebSocket 서버 URL
VITE_WS_BASE_URL=ws://localhost:8080
```

### 환경별 설정

```bash
# 개발 환경
.env.development

# 프로덕션 환경
.env.production

# 로컬 환경 (Git에 커밋되지 않음)
.env.local
```

## 💻 개발 가이드

### 주요 컴포넌트

#### 1. Home.tsx
- 사용자 이름 입력 및 언어 선택
- 방 생성 또는 방 목록으로 이동

#### 2. RoomList.tsx
- 활성화된 게임 방 목록 표시
- 방 참가 기능

#### 3. GameRoom.tsx
- 메인 게임 화면
- 실시간 그리기 및 채팅
- 점수 및 타이머 표시

#### 4. DrawingCanvas.tsx
- HTML5 Canvas 기반 드로잉
- 실시간 그리기 데이터 전송
- 색상 선택 및 캔버스 초기화

### 서비스 레이어

#### API Service (api.ts)
```typescript
// 방 목록 조회
const rooms = await apiService.getGameRooms();

// 방 생성
const result = await apiService.createGameRoom('UserName');

// 방 참가
await apiService.joinGameRoom(roomId, UserName);

// 채팅/정답 제출
const response = await apiService.handleChatMessage(roomId, UserName, message);
```

#### WebSocket Service (websocket.ts)
```typescript
// 연결
await wsService.connect();

// 채널 구독
const subscription = wsService.subscribe('game/room-123', (data) => {
  console.log('Received:', data);
});

// 메시지 전송
wsService.sendMessage(roomId, 'chat', { UserName, message });

// 구독 해제
subscription.unsubscribe();
```

### 타입 정의

모든 타입은 `src/types/index.ts`에 정의되어 있습니다:

- `GameRoom`: 게임 방 데이터
- `Player`: 플레이어 정보
- `ChatMessage`: 채팅 메시지
- `WebSocketRequest/Response`: WebSocket 메시지
- `ApiResult`: API 응답

### 다국어 지원

#### 새 언어 추가

1. `src/i18n/translations.ts`에 번역 추가:
```typescript
export const translations: Record<Language, Translations> = {
  ko: { /* 한국어 */ },
  en: { /* 영어 */ },
  ja: { /* 일본어 */ },
  zh: { /* 중국어 - 새로 추가 */ },
};
```

2. `Language` 타입 업데이트:
```typescript
export type Language = 'en' | 'ko' | 'ja' | 'zh';
```

3. `Home.tsx`에 언어 옵션 추가:
```typescript
const languageOptions = [
  { code: 'ko', label: '한국어' },
  { code: 'en', label: 'English' },
  { code: 'ja', label: '日本語' },
  { code: 'zh', label: '中文' },
];
```

## 🎨 스타일링

현재 프로젝트는 인라인 스타일을 사용합니다. 향후 개선 사항:

### 추천 스타일링 방법
1. **CSS Modules**: 컴포넌트 단위 스타일 격리
2. **Tailwind CSS**: 유틸리티 기반 스타일링
3. **Styled Components**: CSS-in-JS

## 🧪 테스트

### 테스트 프레임워크 설치 (권장)

```bash
# Vitest + Testing Library 설치
npm install --save-dev vitest @testing-library/react @testing-library/jest-dom
```

### 테스트 실행

```bash
npm run test
```

## 📦 빌드 및 배포

### 로컬 빌드

```bash
# 프로덕션 빌드
npm run build

# dist/ 폴더에 빌드 결과 생성
```

### 배포 옵션

#### 1. Vercel
```bash
npm install -g vercel
vercel
```

#### 2. Netlify
```bash
npm install -g netlify-cli
netlify deploy
```

#### 3. GitHub Pages
```bash
# vite.config.ts에 base 설정 추가
base: '/repository-name/'

npm run build
git subtree push --prefix dist origin gh-pages
```

#### 4. Docker
```dockerfile
FROM node:18-alpine as build
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

## 🎯 주요 기능

### 1. 실시간 그리기
- Canvas API를 사용한 부드러운 드로잉
- 색상 팔레트 제공
- 실시간 동기화 (WebSocket)

### 2. 채팅 시스템
- 실시간 채팅
- 정답 자동 검증
- 시스템 메시지 표시

### 3. 게임 진행
- 자동 타이머 관리
- 라운드 진행
- 점수 계산 및 표시

### 4. 다국어 지원
- 한국어, 영어, 일본어 지원
- Context API를 사용한 언어 전환

## 🔍 디버깅

### 개발자 도구 활성화

브라우저 콘솔에서 WebSocket 메시지 확인:

```javascript
// WebSocket 메시지 로깅
wsService.connect(
  () => console.log('WS Connected'),
  (error) => console.error('WS Error:', error)
);
```

### 일반적인 문제 해결

#### 1. WebSocket 연결 실패
```
Error: WebSocket connection to 'ws://localhost:8080/app/ws' failed
```
**해결**: 백엔드 서버가 실행 중인지 확인

#### 2. CORS 오류
```
Access to XMLHttpRequest has been blocked by CORS policy
```
**해결**: 백엔드에서 CORS 설정 확인

#### 3. 타입 오류
```
Property 'xxx' does not exist on type 'yyy'
```
**해결**: `src/types/index.ts`에서 타입 정의 확인

## 📚 추가 리소스

- [React 공식 문서](https://react.dev/)
- [TypeScript 핸드북](https://www.typescriptlang.org/docs/)
- [Vite 가이드](https://vitejs.dev/guide/)
- [Axios 문서](https://axios-http.com/docs/intro)
- [WebSocket API](https://developer.mozilla.org/en-US/docs/Web/API/WebSocket)

## 🤝 기여하기

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📄 라이선스

이 프로젝트는 MIT 라이선스를 따릅니다.

---

Made with ❤️ using React + TypeScript
