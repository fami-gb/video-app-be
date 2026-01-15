# フロントエンド実装方針 (Frontend Implementation Strategy)

## 目的
動画にタグ（例: #動物, #友達）を付けることで、動画を探しやすくする機能を実装します。

## フロントエンド実装の全体像

### 1. UI/UX 設計

#### 1.1 動画アップロード画面の拡張
**場所**: `app/page.tsx` または専用のアップロードコンポーネント

**追加要素**:
```typescript
// タグ入力フィールドの追加（任意入力）
- タグ入力欄 (Input/TagInput コンポーネント)
- タグプレビュー表示エリア
- タグの追加/削除ボタン
```

**推奨UI実装パターン**:

**パターンA: シンプルなテキスト入力**
```tsx
<input 
  type="text" 
  placeholder="タグを入力 (例: 動物, 友達)" 
  value={tagInput}
  onChange={(e) => setTagInput(e.target.value)}
/>
<button onClick={handleAddTag}>タグを追加</button>

{/* タグ表示エリア */}
<div className="tags-container">
  {tags.map((tag, index) => (
    <span key={index} className="tag">
      #{tag}
      <button onClick={() => removeTag(index)}>×</button>
    </span>
  ))}
</div>
```

**パターンB: リアルタイムタグ生成（カンマ区切り）**
```tsx
<input 
  type="text" 
  placeholder="タグをカンマ区切りで入力 (例: 動物, 友達, 旅行)" 
  value={tagsInput}
  onChange={(e) => setTagsInput(e.target.value)}
  onBlur={handleTagsParse} // フォーカスアウト時に配列に変換
/>
```

**パターンC: タグサジェスト付き（発展版）**
```tsx
<TagInput 
  suggestions={popularTags} 
  onTagAdd={handleTagAdd}
  onTagRemove={handleTagRemove}
  maxTags={10}
/>
```

#### 1.2 動画一覧・検索画面の拡張
**場所**: `app/page.tsx` または `components/VideoList.tsx`

**追加要素**:
```typescript
// タグフィルター機能
- タグ検索バー
- タグクラウド表示（人気タグの表示）
- 選択中のタグ表示
- タグでのフィルタリング結果
```

**推奨実装**:
```tsx
// 検索バー
<input 
  type="text"
  placeholder="タグで検索 (例: 動物)"
  value={searchTag}
  onChange={(e) => setSearchTag(e.target.value)}
/>
<button onClick={() => handleSearchByTag(searchTag)}>検索</button>

// または各動画のタグをクリックで検索
<div className="video-tags">
  {video.tags?.map((tag) => (
    <span 
      key={tag} 
      className="tag clickable"
      onClick={() => handleSearchByTag(tag)}
    >
      #{tag}
    </span>
  ))}
</div>
```

### 2. データ管理（State Management）

#### 2.1 アップロード時の状態管理
```typescript
// アップロードフォームのstate例
const [title, setTitle] = useState<string>('');
const [videoFile, setVideoFile] = useState<File | null>(null);
const [tags, setTags] = useState<string[]>([]);  // 新規追加
const [tagInput, setTagInput] = useState<string>('');  // 入力中のタグ

// タグ追加処理
const handleAddTag = () => {
  if (tagInput.trim() && !tags.includes(tagInput.trim())) {
    setTags([...tags, tagInput.trim()]);
    setTagInput('');
  }
};

// タグ削除処理
const removeTag = (index: number) => {
  setTags(tags.filter((_, i) => i !== index));
};
```

#### 2.2 検索時の状態管理
```typescript
const [videos, setVideos] = useState<Video[]>([]);
const [filteredVideos, setFilteredVideos] = useState<Video[]>([]);
const [searchTag, setSearchTag] = useState<string>('');

// タグ検索処理
const handleSearchByTag = async (tag: string) => {
  setSearchTag(tag);
  
  // バックエンドAPIでフィルタリング
  const response = await fetch(`/api/videos?tag=${encodeURIComponent(tag)}`);
  const data = await response.json();
  setFilteredVideos(data);
  
  // またはフロントエンドでフィルタリング
  // const filtered = videos.filter(video => 
  //   video.tags?.includes(tag)
  // );
  // setFilteredVideos(filtered);
};
```

### 3. API連携

#### 3.1 動画アップロード時のAPI呼び出し修正
```typescript
// POST /api/videos のリクエストボディにtagsを追加
const createVideo = async (title: string, videoKey: string, tags: string[]) => {
  const response = await fetch('/api/videos', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      title,
      video_key: videoKey,
      tags,  // 新規追加: タグの配列を送信
    }),
  });
  
  if (!response.ok) {
    throw new Error('Failed to create video');
  }
  
  return response.json();
};
```

#### 3.2 動画一覧取得時のAPI呼び出し
```typescript
// GET /api/videos にタグフィルタパラメータを追加
const fetchVideos = async (tag?: string) => {
  const url = tag 
    ? `/api/videos?tag=${encodeURIComponent(tag)}`
    : '/api/videos';
    
  const response = await fetch(url);
  if (!response.ok) {
    throw new Error('Failed to fetch videos');
  }
  
  return response.json();
};
```

### 4. TypeScript型定義の更新

```typescript
// types/video.ts または適切な場所

// Video型の更新
export interface Video {
  ID: number;
  CreatedAt: string;
  UpdatedAt: string;
  DeletedAt?: string | null;
  title: string;
  url: string;
  video_key: string;
  tags?: string[];  // 新規追加: オプショナルなタグ配列
}

// CreateVideoRequest型
export interface CreateVideoRequest {
  title: string;
  video_key: string;
  tags?: string[];  // 新規追加: オプショナルなタグ配列
}
```

### 5. コンポーネント設計案

#### 推奨コンポーネント分割
```
app/
├── page.tsx (メインページ)
├── components/
│   ├── VideoUpload.tsx (アップロードフォーム)
│   ├── TagInput.tsx (タグ入力コンポーネント)
│   ├── VideoList.tsx (動画一覧表示)
│   ├── VideoCard.tsx (個別動画カード)
│   ├── TagFilter.tsx (タグ検索・フィルター)
│   └── TagCloud.tsx (人気タグ表示 - 発展版)
```

#### TagInputコンポーネント例
```tsx
// components/TagInput.tsx
interface TagInputProps {
  tags: string[];
  onTagAdd: (tag: string) => void;
  onTagRemove: (index: number) => void;
  maxTags?: number;
}

export const TagInput: React.FC<TagInputProps> = ({
  tags,
  onTagAdd,
  onTagRemove,
  maxTags = 10,
}) => {
  const [input, setInput] = useState('');

  const handleAdd = () => {
    const trimmed = input.trim();
    if (trimmed && !tags.includes(trimmed) && tags.length < maxTags) {
      onTagAdd(trimmed);
      setInput('');
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      handleAdd();
    }
  };

  return (
    <div className="tag-input-container">
      <div className="input-wrapper">
        <input
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="タグを入力してEnter"
          maxLength={20}
        />
        <button onClick={handleAdd} disabled={tags.length >= maxTags}>
          追加
        </button>
      </div>
      
      <div className="tags-display">
        {tags.map((tag, index) => (
          <span key={index} className="tag">
            #{tag}
            <button onClick={() => onTagRemove(index)}>×</button>
          </span>
        ))}
        {tags.length >= maxTags && (
          <span className="tag-limit-message">
            最大{maxTags}個まで
          </span>
        )}
      </div>
    </div>
  );
};
```

### 6. バリデーション

#### フロントエンドバリデーション
```typescript
const validateTag = (tag: string): boolean => {
  // タグの検証ルール
  const trimmed = tag.trim();
  
  // 空文字チェック
  if (!trimmed) return false;
  
  // 長さチェック (1-20文字)
  if (trimmed.length < 1 || trimmed.length > 20) return false;
  
  // 特殊文字チェック（必要に応じて）
  const validPattern = /^[a-zA-Z0-9ぁ-んァ-ヶー一-龯]+$/;
  if (!validPattern.test(trimmed)) return false;
  
  return true;
};

const handleAddTag = () => {
  if (!validateTag(tagInput)) {
    alert('タグは1-20文字で、特殊文字は使用できません');
    return;
  }
  
  if (tags.includes(tagInput.trim())) {
    alert('このタグは既に追加されています');
    return;
  }
  
  if (tags.length >= 10) {
    alert('タグは最大10個までです');
    return;
  }
  
  setTags([...tags, tagInput.trim()]);
  setTagInput('');
};
```

### 7. スタイリング推奨

#### CSS例 (Tailwind CSS使用時)
```tsx
// タグ表示スタイル
<span className="inline-flex items-center gap-1 px-3 py-1 bg-blue-100 text-blue-800 rounded-full text-sm">
  #{tag}
  <button 
    className="ml-1 text-blue-600 hover:text-blue-800"
    onClick={onRemove}
  >
    ×
  </button>
</span>

// クリック可能なタグ
<span className="inline-flex items-center px-3 py-1 bg-gray-100 text-gray-800 rounded-full text-sm cursor-pointer hover:bg-gray-200 transition">
  #{tag}
</span>
```

### 8. 実装の優先順位

#### フェーズ1（MVP: 最小限の機能）
1. ✅ タグ入力フィールドの追加（シンプルなテキスト入力）
2. ✅ タグ表示エリア（追加・削除機能）
3. ✅ POST /api/videos 時にタグを送信
4. ✅ 動画一覧でタグを表示

#### フェーズ2（検索機能）
1. ✅ タグ検索バーの追加
2. ✅ GET /api/videos?tag=xxx での検索
3. ✅ タグクリックで検索機能

#### フェーズ3（発展機能）
1. ⬜ タグサジェスト機能
2. ⬜ 人気タグのタグクラウド表示
3. ⬜ 複数タグでのAND/OR検索
4. ⬜ タグの使用回数表示

### 9. テスト観点

#### フロントエンドテスト項目
- [ ] タグの追加/削除が正常に動作する
- [ ] 重複タグを追加できない
- [ ] 最大タグ数制限が機能する
- [ ] タグなしでも動画アップロードできる
- [ ] タグ検索が正常に動作する
- [ ] タグ表示が正しくレンダリングされる
- [ ] 空のタグを追加できない
- [ ] 特殊文字のバリデーションが機能する

### 10. バックエンドAPI仕様（本リポジトリで実装）

#### POST /api/videos
**リクエスト**:
```json
{
  "title": "動画タイトル",
  "video_key": "1234567890-video.mp4",
  "tags": ["動物", "友達", "旅行"]
}
```

**レスポンス**:
```json
{
  "ID": 1,
  "CreatedAt": "2026-01-15T22:00:00Z",
  "UpdatedAt": "2026-01-15T22:00:00Z",
  "DeletedAt": null,
  "title": "動画タイトル",
  "url": "https://example.com/1234567890-video.mp4",
  "video_key": "1234567890-video.mp4",
  "tags": ["動物", "友達", "旅行"]
}
```

#### GET /api/videos
**パラメータなし**: すべての動画を取得

**パラメータあり**: `?tag=動物`
- 指定したタグを含む動画のみを取得

**レスポンス**:
```json
[
  {
    "ID": 1,
    "CreatedAt": "2026-01-15T22:00:00Z",
    "UpdatedAt": "2026-01-15T22:00:00Z",
    "DeletedAt": null,
    "title": "動画タイトル",
    "url": "https://example.com/1234567890-video.mp4",
    "video_key": "1234567890-video.mp4",
    "tags": ["動物", "友達"]
  }
]
```

## まとめ

このフロントエンド実装方針に従うことで、段階的にタグ機能を実装できます。

**推奨実装順序**:
1. 最小限のUI（タグ入力フィールド）を追加
2. バックエンドAPIとの連携を確認
3. タグ表示とフィルタリング機能を追加
4. UI/UXの改善と発展機能の追加

**注意点**:
- タグは任意入力なので、タグなしでも動画アップロード可能にする
- バリデーションはフロントエンド・バックエンド両方で実施
- ユーザビリティを考慮したUIデザインを心がける
