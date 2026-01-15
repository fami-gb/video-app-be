# API Documentation - Video App Backend

## Base URL
- Development: `http://localhost:8080`
- Production: `https://video-app-be.onrender.com`

## Authentication
Currently, no authentication is required for API endpoints.

---

## Endpoints

### 1. Health Check

**GET /**

ヘルスチェック用エンドポイント。バックエンドが正常に動作しているかを確認します。

#### Request
```
GET /
```

#### Response
**Status**: `200 OK`

```json
{
  "message": "Video App Backend is running!"
}
```

---

### 2. Get All Videos

**GET /api/videos**

すべての動画を取得します。オプションでタグによるフィルタリングが可能です。

#### Request
```
GET /api/videos
```

#### Query Parameters
| Parameter | Type   | Required | Description                                    |
|-----------|--------|----------|------------------------------------------------|
| `tag`     | string | No       | 指定したタグを含む動画のみを取得（完全一致） |

#### Examples

**全動画を取得**:
```bash
curl http://localhost:8080/api/videos
```

**タグでフィルタリング**:
```bash
curl http://localhost:8080/api/videos?tag=動物
```

#### Response
**Status**: `200 OK`

```json
[
  {
    "ID": 1,
    "CreatedAt": "2026-01-15T22:00:00Z",
    "UpdatedAt": "2026-01-15T22:00:00Z",
    "DeletedAt": null,
    "title": "猫の動画",
    "url": "https://example.com/1234567890-cat.mp4",
    "video_key": "1234567890-cat.mp4",
    "tags": ["動物", "ペット", "猫"]
  },
  {
    "ID": 2,
    "CreatedAt": "2026-01-15T22:05:00Z",
    "UpdatedAt": "2026-01-15T22:05:00Z",
    "DeletedAt": null,
    "title": "旅行の思い出",
    "url": "https://example.com/1234567891-travel.mp4",
    "video_key": "1234567891-travel.mp4",
    "tags": ["旅行", "友達", "思い出"]
  }
]
```

**Note**: 
- `tags`フィールドが空の場合は空配列`[]`が返されます
- タグが設定されていない古いデータの場合は`null`または空配列になります

---

### 3. Create Video

**POST /api/videos**

動画のメタデータをデータベースに登録します。このエンドポイントは、動画ファイルのアップロードが完了した後に呼び出されます。

#### Request
```
POST /api/videos
Content-Type: application/json
```

#### Request Body
```json
{
  "title": "動画のタイトル",
  "video_key": "1234567890-video.mp4",
  "tags": ["動物", "友達", "旅行"]
}
```

| Field       | Type     | Required | Description                                                      |
|-------------|----------|----------|------------------------------------------------------------------|
| `title`     | string   | Yes      | 動画のタイトル                                                   |
| `video_key` | string   | Yes      | R2ストレージに保存された動画のキー（`/upload-url`で取得したもの）|
| `tags`      | string[] | No       | 動画に関連付けるタグの配列（最大10個推奨、各タグ最大20文字推奨） |

#### Example
```bash
curl -X POST http://localhost:8080/api/videos \
  -H "Content-Type: application/json" \
  -d '{
    "title": "猫の動画",
    "video_key": "1234567890-cat.mp4",
    "tags": ["動物", "ペット", "猫"]
  }'
```

**タグなしの例**:
```bash
curl -X POST http://localhost:8080/api/videos \
  -H "Content-Type: application/json" \
  -d '{
    "title": "無題の動画",
    "video_key": "1234567890-video.mp4",
    "tags": []
  }'
```

#### Response
**Status**: `201 Created`

```json
{
  "ID": 1,
  "CreatedAt": "2026-01-15T22:00:00Z",
  "UpdatedAt": "2026-01-15T22:00:00Z",
  "DeletedAt": null,
  "title": "猫の動画",
  "url": "https://example.com/1234567890-cat.mp4",
  "video_key": "1234567890-cat.mp4",
  "tags": ["動物", "ペット", "猫"]
}
```

#### Error Responses

**Invalid Input** - `400 Bad Request`
```json
{
  "error": "Invalid input"
}
```

**Missing Configuration** - `500 Internal Server Error`
```json
{
  "error": "Public domain configuration missing"
}
```

**Database Error** - `500 Internal Server Error`
```json
{
  "error": "Failed to save video"
}
```

---

### 4. Generate Upload URL

**POST /api/upload-url**

動画アップロード用の署名付きURL（presigned URL）を生成します。このURLを使用して、フロントエンドから直接R2ストレージに動画をアップロードできます。

#### Request
```
POST /api/upload-url
Content-Type: application/json
```

#### Request Body
```json
{
  "filename": "myvideo.mp4"
}
```

| Field      | Type   | Required | Description                     |
|------------|--------|----------|---------------------------------|
| `filename` | string | Yes      | アップロードする動画のファイル名 |

#### Example
```bash
curl -X POST http://localhost:8080/api/upload-url \
  -H "Content-Type: application/json" \
  -d '{"filename": "myvideo.mp4"}'
```

#### Response
**Status**: `200 OK`

```json
{
  "uploadUrl": "https://account-id.r2.cloudflarestorage.com/bucket/1234567890-myvideo.mp4?X-Amz-Algorithm=...",
  "key": "1234567890-myvideo.mp4"
}
```

| Field       | Type   | Description                                                      |
|-------------|--------|------------------------------------------------------------------|
| `uploadUrl` | string | 動画アップロード用の署名付きURL（15分間有効）                     |
| `key`       | string | R2ストレージ上の動画キー（動画登録時に使用）                      |

#### Usage Flow
1. このエンドポイントを呼び出してuploadURLとkeyを取得
2. 取得したuploadURLに対してPUTリクエストで動画ファイルをアップロード
3. アップロード完了後、取得したkeyを使用して`POST /api/videos`で動画情報を登録

#### Error Responses

**Invalid Input** - `400 Bad Request`
```json
{
  "error": "Invalid input"
}
```

**Upload URL Generation Failed** - `500 Internal Server Error`
```json
{
  "error": "Failed to generate upload URL"
}
```

---

### 5. Delete Video

**DELETE /api/videos/:id**

指定したIDの動画を削除します。データベースとR2ストレージの両方から削除されます。

#### Request
```
DELETE /api/videos/:id
```

#### Path Parameters
| Parameter | Type    | Required | Description      |
|-----------|---------|----------|------------------|
| `id`      | integer | Yes      | 削除する動画のID |

#### Example
```bash
curl -X DELETE http://localhost:8080/api/videos/1
```

#### Response
**Status**: `200 OK`

```json
{
  "message": "Video deleted successfully"
}
```

#### Error Responses

**Video Not Found** - `404 Not Found`
```json
{
  "error": "Video not found"
}
```

**Storage Deletion Failed** - `500 Internal Server Error`
```json
{
  "error": "Failed to delete video from storage"
}
```

**Database Deletion Failed** - `500 Internal Server Error`
```json
{
  "error": "Failed to delete video from database"
}
```

---

## Data Models

### Video
```typescript
interface Video {
  ID: number;                    // 動画の一意識別子
  CreatedAt: string;             // 作成日時 (ISO 8601形式)
  UpdatedAt: string;             // 更新日時 (ISO 8601形式)
  DeletedAt: string | null;      // 削除日時 (論理削除、通常はnull)
  title: string;                 // 動画のタイトル
  url: string;                   // 動画の公開URL
  video_key: string;             // R2ストレージ上のキー
  tags: string[];                // タグの配列 (新機能)
}
```

---

## Tag Feature Specifications

### タグの仕様

#### 制限事項
- **タグの最大数**: 制限なし（推奨: 10個まで）
- **タグの最大文字数**: 制限なし（推奨: 20文字まで）
- **タグの形式**: 任意の文字列（フロントエンドでバリデーション推奨）
- **重複タグ**: 許可されています（フロントエンドで重複チェック推奨）

#### タグのフィルタリング
- `GET /api/videos?tag=xxx` でタグによるフィルタリングが可能
- 完全一致検索（部分一致ではない）
- 大文字小文字は区別される
- 複数タグによるフィルタリングは現在サポートされていません

#### データベース実装
- PostgreSQLの`text[]`型を使用して配列として保存
- インデックスの追加によって検索パフォーマンスの向上が可能（将来の改善項目）

---

## Frontend Integration Examples

### Complete Upload Flow with Tags

```typescript
// 1. Generate upload URL
const generateUploadUrl = async (filename: string) => {
  const response = await fetch('http://localhost:8080/api/upload-url', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ filename }),
  });
  return response.json();
};

// 2. Upload video file
const uploadVideo = async (uploadUrl: string, file: File) => {
  const response = await fetch(uploadUrl, {
    method: 'PUT',
    body: file,
  });
  return response.ok;
};

// 3. Register video with tags
const registerVideo = async (
  title: string,
  videoKey: string,
  tags: string[]
) => {
  const response = await fetch('http://localhost:8080/api/videos', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      title,
      video_key: videoKey,
      tags,
    }),
  });
  return response.json();
};

// Complete flow
const handleVideoUpload = async (
  title: string,
  file: File,
  tags: string[]
) => {
  try {
    // Step 1: Get upload URL
    const { uploadUrl, key } = await generateUploadUrl(file.name);
    
    // Step 2: Upload video
    const uploadSuccess = await uploadVideo(uploadUrl, file);
    if (!uploadSuccess) {
      throw new Error('Upload failed');
    }
    
    // Step 3: Register video metadata with tags
    const video = await registerVideo(title, key, tags);
    console.log('Video registered:', video);
    
    return video;
  } catch (error) {
    console.error('Upload error:', error);
    throw error;
  }
};
```

### Fetching Videos with Tag Filter

```typescript
// Get all videos
const fetchAllVideos = async () => {
  const response = await fetch('http://localhost:8080/api/videos');
  return response.json();
};

// Get videos by tag
const fetchVideosByTag = async (tag: string) => {
  const response = await fetch(
    `http://localhost:8080/api/videos?tag=${encodeURIComponent(tag)}`
  );
  return response.json();
};

// Usage
const videos = await fetchAllVideos();
const animalVideos = await fetchVideosByTag('動物');
```

---

## CORS Configuration

The backend allows requests from:
- `http://localhost:3000` (development)
- Environment variable `FRONTEND_URL` (production)

Allowed methods:
- `GET`
- `POST`
- `PUT`
- `DELETE`

---

## Environment Variables

The following environment variables are required:

```bash
# Database Configuration
POSTGRES_USER=user
POSTGRES_PASSWORD=password
POSTGRES_DB=videodb
POSTGRES_HOST=db
DATABASE_URL=postgresql://user:password@db:5432/videodb  # Optional, overrides individual DB vars

# R2 Storage Configuration
R2_ACCOUNT_ID=your-account-id
R2_ACCESS_KEY_ID=your-access-key
R2_SECRET_ACCESS_KEY=your-secret-key
R2_BUCKET_NAME=your-bucket-name

# Public Domain
PUBLIC_DOMAIN=https://your-public-domain.com  # Used for generating video URLs

# Frontend URL (for CORS)
FRONTEND_URL=https://your-frontend-domain.com
```

---

## Migration Notes

### Database Schema Update

新しい`tags`カラムが追加されました。GORMの`AutoMigrate`により自動的にテーブルスキーマが更新されます。

既存のデータベースでは、既存レコードの`tags`フィールドは空の配列またはNULLになります。

### Backward Compatibility

- 既存のフロントエンドコードは、`tags`フィールドを送信しなくても引き続き動作します
- `tags`フィールドが省略された場合、空の配列として保存されます
- 既存のAPIレスポンスには`tags`フィールドが追加されますが、後方互換性があります

---

## Future Enhancements

以下の機能追加が検討されています:

1. **複数タグフィルタリング**: `?tags=動物,ペット` のようなAND/OR検索
2. **タグの部分一致検索**: `?tag_like=動` で「動物」「動画」などを検索
3. **人気タグ取得API**: `GET /api/tags/popular` で使用頻度の高いタグを取得
4. **タグのオートコンプリート**: 既存タグの候補を返すAPI
5. **タグの正規化**: 同じ意味のタグを統一（例: "動物"と"どうぶつ"）
6. **タグ数の制限**: バックエンドでの最大タグ数制限
7. **インデックス最適化**: PostgreSQLのGINインデックスによる検索パフォーマンス向上

---

## Version History

### v1.1.0 (2026-01-15)
- ✨ タグ機能の追加
  - `Video`モデルに`tags`フィールドを追加
  - `POST /api/videos`でタグの受け付けと保存
  - `GET /api/videos?tag=xxx`でタグフィルタリング機能
- 📝 API仕様書の作成

### v1.0.0 (Initial Release)
- 動画のアップロード・取得・削除機能
- R2ストレージ連携
- PostgreSQLデータベース連携
