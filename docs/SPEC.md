# InterKnot ゲーム仕様

## ゲーム概要

GPS で取得した現実の位置情報をもとにプレイヤー同士をループ状に繋ぎ、隣接プレイヤーとのチャットや位置交換によって線の交差を解消していく協力型リアルタイム Web ゲーム。

---

## ゲームの流れ

### 1. マッチング

- 名前と GPS 座標を送信して `/api/rooms/join` を叩くと、待機中のルームに入る (なければ新規作成)。
- **定員: 8 人**。8 人揃うと 5 秒後にゲームが自動開始する。
- 参加時にアカウント作成・ログインは不要。サーバーが発行する `player_id` (UUID) をクライアントが保持し、以降のリクエストには `X-Player-ID` ヘッダで渡す。

### 2. ゲーム開始

- ゲーム開始時、各プレイヤーに `order_index` (0 〜 n-1) が割り振られる。
- `order_index` の割り振りはランダムシャッフルで行うが、以下の条件を満たすまで最大 100 回やり直す。
  - 交差数が 1 以上であること (= 解消すべき絡まりが存在する)
  - 任意の 1 スワップだけでは解けないこと (= 最低 2 スワップ必要)
- GPS 座標はゲーム開始時に 1 度だけ取得し、以降変化しない。
- **制限時間: 5 分**。

### 3. プレイ中

- プレイヤーは `order_index` 順にループ状に繋がる閉路グラフを形成する。
  - 0 → 1 → 2 → … → n-1 → 0 の順で辺が張られる。
- 隣接プレイヤー (`order_index` の差が 1、または 0 と n-1 の折り返し) とのみチャット・スワップが可能。

#### チャット

- 隣接プレイヤーにのみメッセージを送信できる。
- メッセージは 1 〜 2000 文字。
- 過去に隣接していた相手のチャット履歴はゲーム終了まで閲覧できる。

#### スワップ (order_index の交換)

1. 自分が相手に対して `PUT /api/rooms/{room_id}/connections/{target_id}` を `{ "needs_swap": true }` で叩き、交換の意思を表明する。
2. 相手も同じ操作で意思表明する。
3. 双方の `needs_swap` が `true` になった瞬間にサーバーが自動的に 2 人の `order_index` を交換する (= グラフ上の繋がりが入れ替わる)。
4. スワップ成立後、その 2 人が関わる `connections` レコードはすべて削除される (隣接相手が変わるため)。
5. `{ "needs_swap": false }` を送ると意思表明を取り下げられる。

> **注意:** 相手が自分のスワップ申請を知る手段はない。チャットで合図を取り合う必要がある。

### 4. 終了条件

| 結果 | 条件 |
|------|------|
| クリア (finished) | 交差数が 0 になった瞬間 |
| タイムアウト (finished) | 制限時間 5 分を超えた |

終了後の全体チャットフェーズはない。

---

## 交差判定

- データベース側: PostGIS の `ST_Crosses` 関数で辺ペアの交差を判定する。
- メモリ内 (シャッフル時): 平面クロス積による線分交差判定を使用する。
- 隣接する辺 (端点を共有する辺) は交差カウントから除外する。

---

## SSE イベント

クライアントは `GET /api/rooms/{room_id}/stream` に接続し続けることでリアルタイムに状態変化を受け取る。

| イベント名 | 発生タイミング |
|-----------|--------------|
| `room_updated` | 誰かが参加した / スワップが成立した / タイムアウトした |
| `room_started` | ゲームが開始した |

`room_updated` のペイロードにはスワップ成立時に最新の交差辺ペア一覧 (`intersections`) が含まれる。

---

## API エンドポイント一覧

| Method | Path | 説明 |
|--------|------|------|
| POST | `/api/rooms/join` | マッチングキューに参加 |
| GET | `/api/rooms/{room_id}` | ルーム状態取得 |
| GET | `/api/rooms/{room_id}/stream` | SSE ストリーム購読 |
| GET | `/api/rooms/{room_id}/intersections` | 交差辺ペア一覧取得 |
| PUT | `/api/rooms/{room_id}/connections/{target_id}` | スワップ意思表示 / 取り下げ |
| GET | `/api/rooms/{room_id}/messages` | メッセージ履歴取得 |
| POST | `/api/rooms/{room_id}/messages` | メッセージ送信 |

詳細なリクエスト・レスポンス定義は [`backend/docs/openapi/openapi-v1.yaml`](../backend/docs/openapi/openapi-v1.yaml) を参照。

---

## データモデル

### rooms

| カラム | 型 | 説明 |
|--------|----|------|
| `id` | UUID (PK) | ルーム ID |
| `status` | TEXT | `matching` / `playing` / `finished` |
| `start_at` | TIMESTAMPTZ | ゲーム開始時刻 (開始前は NULL) |
| `expires_at` | TIMESTAMPTZ | 制限時間切れ時刻 |

### players

| カラム | 型 | 説明 |
|--------|----|------|
| `id` | UUID (PK) | プレイヤー ID |
| `room_id` | UUID (FK) | 所属ルーム |
| `name` | VARCHAR(255) | 表示名 |
| `location` | geography(Point, 4326) | GPS 座標 (ゲーム開始時に固定) |
| `order_index` | INT | ループ上の順番 (0 〜 n-1、ルーム内でユニーク) |

### messages

| カラム | 型 | 説明 |
|--------|----|------|
| `id` | UUID (PK) | メッセージ ID |
| `room_id` | UUID (FK) | 所属ルーム |
| `sender_id` | UUID (FK) | 送信者 |
| `receiver_id` | UUID (FK) | 受信者 (NULL の場合は全体チャット、現在未使用) |
| `content` | TEXT | 本文 (1 〜 2000 文字) |
| `created_at` | TIMESTAMPTZ | 送信日時 |

### connections

スワップの意思表示を管理するテーブル。スワップ成立時は関係する 2 人のレコードがすべて削除される。

| カラム | 型 | 説明 |
|--------|----|------|
| `room_id` | UUID (FK) | 所属ルーム |
| `sender_id` | UUID (FK, PK) | 意思表示した側 |
| `receiver_id` | UUID (FK, PK) | 交換相手として指名された側 |
| `needs_swap` | BOOLEAN | `true` = 交換したい |
