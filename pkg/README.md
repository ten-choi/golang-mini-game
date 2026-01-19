# Public Packages

이 디렉토리는 외부 프로젝트에서도 import 가능한 공용 라이브러리를 포함합니다.

## 디렉토리 구조

- `constants/` - 애플리케이션 상수 (게임 타입, 상태 등)
- `utils/` - 유틸리티 함수 (문자열 처리, ID 생성 등)

## 사용 예시

```go
import (
    "draw-and-guess-server/pkg/constants"
    "draw-and-guess-server/pkg/utils"
)

func example() {
    // 상수 사용
    gameType := constants.GameTypeWordChain
    
    // 유틸리티 함수 사용
    id, _ := utils.GenerateID(16)
    isKorean := utils.IsKorean("안녕하세요")
}
```

## 주의사항

- `pkg/` 패키지는 외부 공개 API이므로 breaking changes를 피하세요
- private 코드는 `internal/`에 위치해야 합니다
- 의존성을 최소화하세요 (표준 라이브러리 우선)
