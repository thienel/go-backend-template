---
description: Hướng dẫn quản lý Cơ sở dữ liệu với Ent và Atlas
---
# Hướng dẫn quy trình phát triển Database với Ent và Atlas

Dự án này sử dụng [Ent](https://entgo.io/) làm ORM và [Atlas](https://atlasgo.io/) làm công cụ quản lý Migration chuyên nghiệp. Ent code gen và migration không tự động diễn ra lúc server boot (tính năng AutoMigrate của GORM đã bị xóa để tối ưu High Scalability). Bạn cần quản lý schema và migration một cách rõ ràng. Dưới đây là quy trình chi tiết:

## Bước 1: Tạo Schema mới cho Ent
Nếu bạn muốn thêm một Entity/Bảng mới vào Database (ví dụ bảng `Ticket`):
```bash
make ent-new NAME=Ticket
```
Lệnh này sẽ tạo ra một file `internal/ent/schema/ticket.go`. Bạn sẽ dùng file này để định nghĩa cấu trúc (Fields) và các mối quan hệ (Edges) cho Ticket.

## Bước 2: Sinh mã nguồn (Generate Code)
Sau khi định nghĩa (hoặc thay đổi) các file `.go` trong `internal/ent/schema`, bạn **phải** sinh lại mã nguồn ORM của Ent:

// turbo
```bash
make ent-gen
```
> [!NOTE] Cập nhật lại Dependency
> Đôi khi sau khi sinh mã nguồn, Go module có thể phàn nàn về code mới. Bạn nên chạy thêm `go mod tidy` để fix. Đừng quên reload các file đang mở trong IDE.

## Bước 3: Tạo File Migration bằng Atlas
Sau khi sinh mã nguồn Ent xong, mã Go của bạn đã biết về cấu trúc mới, nhưng Database thật thì chưa thay đổi. Bạn cần Atlas so sánh schema hiện tại của Ent với thư mục Migration (hoặc môi trường Dev Database tạm) để sinh ra file SQL.

**Ghi chú:** Bạn cần cài đặt Atlas CLI trước. Hướng dẫn cài đặt có tại `https://atlasgo.io/getting-started/`.

Chạy lệnh để sinh ra migration (Ví dụ cho Ent):
```bash
atlas migrate diff migration_name \
  --dir "file://internal/ent/migrate/migrations" \
  --to "ent://internal/ent/schema" \
  --dev-url "docker://postgres/15/test"
```
> [!TIP]
> Lưu ý, Atlas yêu cầu một Dev Database trống để nó có thể dry-run cấu trúc Ent sinh ra. Tham số `--dev-url "docker://postgres/15/test"` sẽ tự động tải một Docker container Postgres 15 tạm thời (nếu bạn có Docker). File `.sql` sẽ được tạo ra trong `internal/ent/migrate/migrations`.

## Bước 4: Áp dụng Migration lên Database (Thực thi Migration)
Khi Migration script đã được tạo, để thực sự áp dụng nó lên Database phát triển của bạn (`DB_HOST`, `DB_PORT`,...):

```bash
atlas migrate apply \
  --dir "file://internal/ent/migrate/migrations" \
  --url "postgres://username:password@localhost:5432/go_backend_template?search_path=public&sslmode=disable"
```
*(Lưu ý: Bạn phải lấy đúng chuỗi kết nối PostgreSQL tương ứng với biến môi trường `.env` của dự án).*

## Luồng công việc hàng ngày (Daily Workflow)
1. Thêm Column/Đổi thuộc tính trong file: `internal/ent/schema/[tên].go`
2. Chạy: `make ent-gen` (để sinh mã code ORM).
3. Chạy lệnh: `atlas migrate diff [mô\_tả\_thay\_đổi] ...` (để tạo file SQL versioned).
4. Review lại file `.sql` sinh ra để đảm bảo không bị lỗi sai khác.
5. Chạy lệnh: `atlas migrate apply ...` (áp dụng vào Database local).
6. Commit cả thư mục `internal/ent` và thư mục `internal/ent/migrate/migrations` lên Git.
