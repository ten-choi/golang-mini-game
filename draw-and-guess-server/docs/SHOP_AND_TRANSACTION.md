# 상점 및 거래 시스템

게임 내 아이템 상점과 크레딧 거래 내역을 관리하는 시스템입니다.

## 📦 모델 구조

### ShopItem (상점 아이템)
```go
type ShopItem struct {
    ID           primitive.ObjectID  // 아이템 ID
    Name         string              // 아이템 이름
    Description  string              // 설명
    ItemType     ItemType            // 타입 (avatar/emoji/background/frame/badge/theme)
    Rarity       ItemRarity          // 등급 (common/uncommon/rare/epic/legendary)
    Price        int                 // 가격 (크레딧)
    ImageURL     string              // 이미지 URL
    IsAvailable  bool                // 판매 가능 여부
    IsFeatured   bool                // 추천 상품 여부
    Stock        int                 // 재고 (-1: 무제한)
    RequireLevel int                 // 필요 레벨
    Tags         []string            // 검색 태그
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

### UserInventory (사용자 인벤토리)
```go
type UserInventory struct {
    ID          primitive.ObjectID  // 인벤토리 ID
    UserID      primitive.ObjectID  // 사용자 ID
    ItemID      primitive.ObjectID  // 아이템 ID
    ItemName    string              // 아이템 이름 (캐시)
    ItemType    ItemType            // 아이템 타입 (캐시)
    Quantity    int                 // 수량
    IsEquipped  bool                // 장착 여부
    PurchasedAt time.Time           // 구매 시간
}
```

### Transaction (거래 내역)
```go
type Transaction struct {
    ID            primitive.ObjectID  // 거래 ID
    UserID        primitive.ObjectID  // 사용자 ID
    Type          TransactionType     // 거래 타입
    Status        TransactionStatus   // 거래 상태
    Amount        int                 // 금액 (양수: 획득, 음수: 차감)
    BalanceBefore int                 // 거래 전 잔액
    BalanceAfter  int                 // 거래 후 잔액
    Description   string              // 거래 설명
    Metadata      map[string]interface{} // 추가 데이터
    CreatedAt     time.Time
}
```

## 🎯 주요 기능

### 1. 상점 관리

#### 아이템 조회
```go
// 특정 아이템 조회
item, err := shopService.GetItem(ctx, itemID)

// 모든 아이템 조회 (필터 지원)
filter := &repository.ShopItemFilter{
    ItemType: &models.ItemTypeAvatar,
    IsAvailable: &true,
    MinPrice: &100,
    MaxPrice: &1000,
}
items, err := shopService.GetAllItems(ctx, filter)

// 추천 상품 조회
featured, err := shopService.GetFeaturedItems(ctx)

// 타입별 조회
avatars, err := shopService.GetItemsByType(ctx, models.ItemTypeAvatar)
```

#### 아이템 생성/수정
```go
// 새 아이템 생성
item := &models.ShopItem{
    Name:        "멋진 아바타",
    Description: "레어 등급의 멋진 아바타",
    ItemType:    models.ItemTypeAvatar,
    Rarity:      models.RarityRare,
    Price:       500,
    ImageURL:    "https://example.com/avatar.png",
    IsAvailable: true,
    IsFeatured:  false,
    Stock:       -1, // 무제한
    RequireLevel: 5,
    Tags:        []string{"avatar", "rare", "special"},
}
created, err := shopService.CreateItem(ctx, item)

// 아이템 수정
updated, err := shopService.UpdateItem(ctx, itemID, item)

// 아이템 삭제
err := shopService.DeleteItem(ctx, itemID)
```

### 2. 구매 및 인벤토리

#### 아이템 구매
```go
// 아이템 구매 (잔액 확인, 재고 차감, 거래 생성 자동 처리)
err := shopService.PurchaseItem(ctx, userID, itemID, quantity)
```

#### 인벤토리 조회
```go
// 전체 인벤토리 조회
inventory, err := shopService.GetUserInventory(ctx, userID, nil)

// 특정 타입만 조회
avatarType := models.ItemTypeAvatar
avatars, err := shopService.GetUserInventory(ctx, userID, &avatarType)
```

#### 아이템 장착
```go
// 아이템 장착 (같은 타입의 다른 아이템은 자동 해제)
err := shopService.EquipItem(ctx, userID, itemID)

// 아이템 장착 해제
err := shopService.UnequipItem(ctx, userID, itemID)
```

### 3. 거래 내역 관리

#### 거래 조회
```go
// 특정 거래 조회
tx, err := transactionService.GetTransaction(ctx, transactionID)

// 사용자 거래 내역 조회 (필터 지원)
filter := &repository.TransactionFilter{
    Type: &models.TransactionTypePurchase,
    Status: &models.TransactionStatusCompleted,
    Limit: 20,
    Skip: 0,
}
transactions, err := transactionService.GetUserTransactions(ctx, userID, filter)

// 현재 잔액 조회
balance, err := transactionService.GetUserBalance(ctx, userID)
```

#### 크레딧 지급/차감
```go
// 크레딧 지급
err := transactionService.AddCredit(
    ctx,
    userID,
    500, // 금액
    models.TransactionTypeDailyBonus,
    "Daily login bonus",
    map[string]interface{}{"day": 7},
)

// 크레딧 차감
err := transactionService.DeductCredit(
    ctx,
    userID,
    200,
    models.TransactionTypePurchase,
    "Item purchase",
    nil,
)
```

#### 게임 보상 지급
```go
// 게임 승리 보상
err := transactionService.GrantGameReward(
    ctx,
    userID,
    1000, // 보상 금액
    gameRoomID,
    "WORDCHAIN",
    1, // 순위
    true, // 승리 여부
)
```

## 🔍 거래 타입 (TransactionType)

- `purchase`: 상점 구매
- `reward`: 게임 보상
- `daily_bonus`: 출석 보너스
- `gift`: 선물 받음
- `refund`: 환불
- `admin_grant`: 관리자 지급
- `level_up`: 레벨업 보상
- `achievement`: 업적 보상

## 🏷️ 아이템 타입 (ItemType)

- `avatar`: 아바타
- `emoji`: 이모지
- `background`: 배경
- `frame`: 프레임
- `badge`: 뱃지
- `theme`: 테마

## ⭐ 아이템 등급 (ItemRarity)

- `common`: 일반
- `uncommon`: 고급
- `rare`: 희귀
- `epic`: 영웅
- `legendary`: 전설

## 📊 MongoDB 컬렉션 및 인덱스

### shop_items 컬렉션
- `item_type`: 아이템 타입별 조회
- `is_available`: 판매 가능 여부
- `is_featured`: 추천 상품
- `price`: 가격순 정렬
- `tags`: 태그 검색

### user_inventory 컬렉션
- `user_id + item_id`: 유니크 인덱스 (중복 방지)
- `user_id + item_type`: 타입별 조회
- `user_id + is_equipped`: 장착된 아이템 조회

### transactions 컬렉션
- `user_id + created_at`: 시간순 거래 내역
- `user_id + type`: 타입별 거래 조회
- `user_id + status`: 상태별 거래 조회
- `created_at`: 전체 거래 시간순 정렬

## 💡 사용 예제

### 초기화 (routes/router.go)
```go
// 컬렉션 가져오기
shopItemsCollection := database.GetCollection("shop_items")
userInventoryCollection := database.GetCollection("user_inventory")
transactionsCollection := database.GetCollection("transactions")
usersCollection := database.GetCollection("users")

// Repository 생성
shopRepo := repository.NewShopRepository(shopItemsCollection, userInventoryCollection)
transactionRepo := repository.NewTransactionRepository(transactionsCollection)
userRepo := repository.NewUserRepository(usersCollection)

// Service 생성
shopService := service.NewShopService(shopRepo, transactionRepo, userRepo)
transactionService := service.NewTransactionService(transactionRepo)
```

### 전체 흐름 예제
```go
// 1. 상점 아이템 등록
item := &models.ShopItem{
    Name:         "골든 크라운 아바타",
    Description:  "전설 등급의 황금 왕관 아바타",
    ItemType:     models.ItemTypeAvatar,
    Rarity:       models.RarityLegendary,
    Price:        2000,
    IsAvailable:  true,
    IsFeatured:   true,
    Stock:        100,
    RequireLevel: 10,
}
created, _ := shopService.CreateItem(ctx, item)

// 2. 사용자가 아이템 구매
err := shopService.PurchaseItem(ctx, userID, created.ID.Hex(), 1)
if err != nil {
    // 잔액 부족, 재고 부족, 레벨 부족 등의 에러 처리
}

// 3. 구매한 아이템 장착
err = shopService.EquipItem(ctx, userID, created.ID.Hex())

// 4. 거래 내역 확인
transactions, _ := transactionService.GetUserTransactions(ctx, userID, nil)

// 5. 게임 종료 후 보상 지급
err = transactionService.GrantGameReward(ctx, userID, 500, roomID, "OX", 1, true)
```

## ⚠️ 주의사항

1. **동시성 처리**: 재고 차감 시 race condition 발생 가능 → MongoDB 트랜잭션 사용 권장
2. **잔액 동기화**: Transaction의 `BalanceAfter`와 User 모델의 `Credit`은 별도로 관리됨
3. **재고 관리**: `Stock = -1`은 무제한 재고를 의미
4. **장착 시스템**: 같은 타입의 아이템은 하나만 장착 가능 (자동 해제)
5. **거래 취소**: 구매 실패 시 거래 상태가 `cancelled`로 변경됨

## 🚀 향후 개선 사항

- [ ] MongoDB 트랜잭션 적용 (ACID 보장)
- [ ] User 모델과 Transaction 잔액 동기화
- [ ] 아이템 선물하기 기능
- [ ] 아이템 판매/환불 기능
- [ ] 기간 한정 할인 이벤트
- [ ] 번들 상품 (여러 아이템 묶음 판매)
- [ ] 가챠 시스템
