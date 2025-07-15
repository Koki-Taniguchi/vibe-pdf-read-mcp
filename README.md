# vibe-pdf-read-mcp

ImageMagickを使用してPDFファイルをPNG画像(base64)に変換しLLMが読み込めるようにするModel Context Protocol (MCP) サーバーです。

## 機能

- PDFファイルをPNG画像に変換（品質と解像度のカスタマイズ可能）
- 特定のページまたは全ページの変換
- PDFファイルの総ページ数を取得
- Base64エンコードされた出力で簡単な統合
- AIアシスタントで使用可能なMCPプロトコル準拠

## 必要な環境

- Go 1.24以上
- ImageMagick

## インストール

1. このリポジトリをクローンします：
```bash
git clone https://github.com/Koki-Taniguchi/vibe-pdf-read-mcp.git
cd vibe-pdf-read-mcp
```

2. 実行ファイルをビルドします：
```bash
go build -o vibe-pdf-read-mcp
```

3. ImageMagickがインストールされていない場合はインストールします：

[ImageMagick公式サイト](https://imagemagick.org/script/download.php)

## MCPサーバーとしての使用方法

### 利用可能なツール

#### `convert_pdf_to_images`
PDFファイルをPNG画像に変換します。

**パラメータ:**
- `pdfPath` (必須): 変換するPDFファイルのパス
- `density` (オプション): 解像度（DPI）（デフォルト: 300）
- `quality` (オプション): 画像品質 1-100（デフォルト: 100）
- `page` (オプション): 特定のページ番号（1から始まる）。0または未指定で全ページ変換

**レスポンス:**
- Base64エンコードされたPNG画像データ
- MCP ImageContentタイプとして画像を返却

#### `get_pdf_page_count`
PDFファイルの総ページ数を取得します。

**パラメータ:**
- `pdfPath` (必須): PDFファイルのパス

**レスポンス:**
- PDFの総ページ数

### 設定例

MCPクライアント（例: Claude Desktop）の設定に追加：

```json
{
  "mcpServers": {
    "pdf-to-image": {
      "command": "/path/to/vibe-pdf-read-mcp"
    }
  }
}
```

## トラブルシューティング

### PDF変換が失敗する場合
- ImageMagickがインストールされているか確認: `convert -version`
- ImageMagickセキュリティポリシーを確認（上記参照）
- PDFファイルが存在し、読み取り可能であることを確認

### 大きなPDFでメモリ問題が発生する場合
- `page`パラメータを使用して特定のページのみ変換
- `density`パラメータを下げて解像度を低くする
- PDFをバッチで処理する

## 開発者情報

### 依存関係
- [modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk) - Go用MCP SDK
- ImageMagick - PDF処理用の外部依存関係

### ソースからのビルド
```bash
go mod download
go build -o vibe-pdf-read-mcp
```

### テストの実行
```bash
go test ./...
```

## ライセンス

このプロジェクトはMITライセンスの下でライセンスされています - 詳細は[LICENSE.md](LICENSE.md)ファイルを参照してください。
