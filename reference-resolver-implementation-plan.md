# Reference Resolver - Implementation Plan

## 1. Mục tiêu

Xây dựng cơ chế `Reference Resolver` dùng chung để validate các field tham chiếu (`ref field`) giữa các Model.

Hệ thống hỗ trợ:

- Local Repository
- gRPC Service
- HTTP Service
- Mở rộng resolver khác trong tương lai
- Generic Hook tự động validate reference
- Model không phụ thuộc Repository, gRPC hoặc HTTP
- Không tạo DB connection / Repository mới bên trong Hook
- Không tạo circular dependency giữa Hook, Repository và Resolver

## 2. Kiến trúc

```text
Model (ref:"cruise")
        ↓
Generic Reference Hook
        ↓
DependencyResolver
        ↓
Reference Registry
        ↓
┌───────────┬───────────┬───────────┐
│ Local     │ gRPC      │ HTTP      │
↓           ↓           ↓
Repository  Client      Client
↓           ↓           ↓
MongoDB     Remote      Remote
```

## 3. Package Structure

```text
pkg/
└── resolver/
    ├── resolver.go
    ├── registry.go
    ├── local.go
    ├── grpc.go
    ├── http.go
    ├── extractor.go
    ├── validator.go
    ├── errors.go
    └── resolver_test.go
```

## 4. Resolver Interface

```go
type Resolver interface {
    Exists(ctx context.Context, id string) (bool, error)
}
```

## 5. Generic Repository

```go
type Repository interface {
    ExistsByID(ctx context.Context, id string) (bool, error)
}
```

Repository chỉ chịu trách nhiệm persistence local.

## 6. Local Resolver

```go
type LocalResolver struct {
    repo Repository
}

func Local(repo Repository) Resolver {
    return &LocalResolver{repo: repo}
}

func (r *LocalResolver) Exists(ctx context.Context, id string) (bool, error) {
    return r.repo.ExistsByID(ctx, id)
}
```

## 7. gRPC Resolver

```go
type GRPCClient interface {
    Exists(ctx context.Context, id string) (bool, error)
}

type GRPCResolver struct {
    client GRPCClient
}

func GRPC(client GRPCClient) Resolver {
    return &GRPCResolver{client: client}
}

func (r *GRPCResolver) Exists(ctx context.Context, id string) (bool, error) {
    return r.client.Exists(ctx, id)
}
```

## 8. HTTP Resolver

```go
type HTTPClient interface {
    Exists(ctx context.Context, id string) (bool, error)
}

type HTTPResolver struct {
    client HTTPClient
}

func HTTP(client HTTPClient) Resolver {
    return &HTTPResolver{client: client}
}
```

## 9. Registry

```go
type Registry struct {
    resolvers map[string]Resolver
}

func NewRegistry() *Registry {
    return &Registry{resolvers: make(map[string]Resolver)}
}

func (r *Registry) Register(name string, resolver Resolver) {
    r.resolvers[name] = resolver
}

func (r *Registry) Resolve(name string) (Resolver, error) {
    resolver, ok := r.resolvers[name]
    if !ok {
        return nil, ErrResolverNotFound
    }
    return resolver, nil
}
```

## 10. Model Reference

```go
type Schedule struct {
    ID       string `bson:"_id" json:"id"`
    RouteID  string `bson:"route_id" json:"route_id" ref:"route,required"`
    CruiseID string `bson:"cruise_id" json:"cruise_id" ref:"cruise,required"`
    SiteID   string `bson:"site_id" json:"site_id" ref:"site"`
}
```

Không hard-code transport:

```text
KHÔNG: ref:"cruise.grpc"
KHÔNG: ref:"cruise.http"
KHÔNG: ref:"cruise.local"

ĐÚNG: ref:"cruise"
```

## 11. Reference Structure

```go
type Reference struct {
    Field    string
    Name     string
    ID       string
    Required bool
}
```

## 12. Extractor

```go
func ExtractReferences(model any) ([]Reference, error)
```

Extractor dùng reflection để đọc `ref` tag.

## 13. DependencyResolver

```go
type DependencyResolver struct {
    registry *Registry
}

func NewDependencyResolver(registry *Registry) *DependencyResolver {
    return &DependencyResolver{registry: registry}
}
```

Flow:

```text
Model
 ↓
ExtractReferences
 ↓
DependencyResolver
 ↓
Registry.Resolve
 ↓
Resolver.Exists
```

## 14. Hook Integration

```text
Insert
 ↓
BeforeInsert
 ↓
DependencyResolver.Resolve()
 ↓
Success → Insert DB
Failure → Return Error
```

Hook không được gọi:

```go
datastore.GetConnect()
NewRepository(...)
```

## 15. Dependency Injection

```go
registry := resolver.NewRegistry()

registry.Register("route", resolver.Local(routeRepo))
registry.Register("cruise", resolver.GRPC(cruiseClient))
registry.Register("site", resolver.HTTP(siteClient))

dependencyResolver := resolver.NewDependencyResolver(registry)
```

## 16. Dependency Direction

```text
Hook
 ↓
DependencyResolver
 ↓
Registry
 ↓
Resolver
 ↓
Repository / Client
 ↓
Database / Remote Service
```

Không được tạo dependency ngược.

## 17. Error Handling

Các error code:

```text
REFERENCE_NOT_FOUND
REFERENCE_REQUIRED
REFERENCE_TIMEOUT
REFERENCE_SERVICE_UNAVAILABLE
REFERENCE_RESOLVER_NOT_FOUND
```

Phải phân biệt record không tồn tại với remote service timeout/unavailable.

## 18. Cross-Service Consistency

Remote `Exists()` không phải database foreign key.

Nó chỉ đảm bảo reference tồn tại tại thời điểm validation.

Nếu cần consistency mạnh hơn, xem xét:

- Transaction
- Transactional Outbox
- Event
- Local Projection
- Soft Delete
- Referential Lifecycle Policy

## 19. Resolve và hydrate dữ liệu vào model trước khi save

### Mục tiêu

Một lần gọi resolver phải làm được cả hai việc:

1. Kiểm tra ID reference có tồn tại hay không.
2. Lấy dữ liệu reference và gán vào một field đích của model (nếu model yêu cầu).

Nhờ vậy không cần gọi thêm một hàm query/check ngược lại sau khi validation.

```text
Schedule.RouteID
    ↓
parse ref tag
    ↓
Reference { Name, ID, Field, Target }
    ↓
Route resolver: find by ID
    ↓
not found → trả lỗi, không save
found → trả Route data
    ↓
gán Route data vào Schedule.Route
    ↓
BeforeInsert / BeforeSave hoàn thành
    ↓
save Schedule đã được enrich
```

### Khai báo model đề xuất

`into` chỉ rõ field sẽ nhận dữ liệu được resolver trả về.

```go
type Schedule struct {
    RouteID string `bson:"route_id" json:"route_id" ref:"route,required,into=Route"`
    Route   Route `bson:"route,omitempty" json:"route,omitempty"`

    CruiseID string  `bson:"cruise_id" json:"cruise_id" ref:"cruise,required,into=Cruise"`
    Cruise   Cruise `bson:"cruise,omitempty" json:"cruise,omitempty"`
}
```

Nếu chỉ cần `Route` để xử lý nghiệp vụ hoặc trả API, không muốn lưu snapshot vào DB:

```go
Route Route `bson:"-" json:"route,omitempty"`
```

Nếu cần lưu snapshot Route/Cruise tại thời điểm tạo Schedule thì dùng `bson:"route,omitempty"` như ví dụ phía trên.

### Mở rộng Reference

```go
type Reference struct {
    Field    string // field chứa ID, ví dụ RouteID
    Name     string // tên resolver, ví dụ route
    ID       string // giá trị ID cần resolve
    Required bool
    Target   string // field nhận Message.Data, ví dụ Route
}
```

### Quy ước tag

```go
ref:"route"                    // chỉ kiểm tra tồn tại
ref:"route,required"           // bắt buộc có RouteID, chỉ kiểm tra tồn tại
ref:"route,into=Route"         // kiểm tra và gán data vào Route
ref:"route,required,into=Route" // bắt buộc, kiểm tra và gán data
```

Parser phải từ chối các trường hợp sau:

```go
ref:"route,into="        // target rỗng
ref:"route,into=Unknown" // field target không tồn tại hoặc không set được
ref:"route,into=RouteID" // target không tương thích với dữ liệu resolver trả về
```

### Interface resolver

Resolver nhận pointer model và `prefix` để thực hiện phase kiểm tra/query đúng table/tenant. Việc index và apply result là phase riêng, luôn làm trực tiếp trên pointer model và không dùng `prefix`.

```go
type ResolveFunc func(
    ctx context.Context,
    data any,
    prefix string,
) types.Message
```

Ví dụ resolver local:

```go
registry.Register("route", func(ctx context.Context, data any, prefix string) types.Message {
    schedule := data.(*Schedule)

    route, err := routeRepo.FindByID(ctx, schedule.RouteID, prefix)
    if err != nil {
        return types.Message{
            Status:     "error",
            Code:       404,
            Message:    "route not found",
            MessageErr: err,
        }
    }

    return types.Message{
        Status: "success",
        Code:   200,
        Data:   route,
    }
})
```

`Message.Data` là dữ liệu đã resolve. Với `Target == ""`, validator chỉ dùng kết quả success/error để validate và bỏ qua `Data`.

### Dependency: index field đích và apply dữ liệu vào pointer hiện tại

`Dependency` không tạo model mới và cũng không query lần thứ hai. Nó nhận chính pointer của struct đang được repository xử lý, sau đó dùng `Reference.Target` để tìm (index) field đích và set data resolver vừa trả về.

API đề xuất:

```go
type Dependency struct {
    model reflect.Value // pointer tới model hiện tại, ví dụ *Schedule
}

func NewDependency(model any) (*Dependency, error)
func (d *Dependency) Apply(ref Reference, data any) error
```

`NewDependency` validate một lần rằng `model` là non-nil pointer tới struct. `Apply` dùng `ref.Target` để index field bằng `FieldByName`, sau đó decode/assign `data` vào field đó.

Pseudo-code:

```go
func (d *Dependency) Apply(ref Reference, data any) error {
    if ref.Target == "" || data == nil {
        return nil // ref chỉ có nhiệm vụ validate tồn tại
    }

    target := d.model.Elem().FieldByName(ref.Target)
    if !target.IsValid() || !target.CanSet() {
        return fmt.Errorf("reference target %q is not settable", ref.Target)
    }

    return decodeAndSet(target, data)
}
```

Quy tắc của `Dependency`:

1. `model` phải là pointer tới struct.
2. `target` phải là exported field tồn tại và set được.
3. Kiểu `data` phải assignable/convertible vào kiểu field đích; với map có thể decode sang struct đích.
4. Chỉ field `Target` được thay đổi, không merge tùy ý toàn bộ `Message.Data` vào model.
5. Nếu apply thất bại, trả lỗi và không gọi DB create/update.

Ví dụ:

```go
schedule := &Schedule{RouteID: "route-1"}
// resolver trả về &Route{ID: "route-1", Name: "Ha Long"}
// dependency.Apply(Reference{Target: "Route"}, message.Data)
// schedule.Route.Name == "Ha Long"
```

### Flow trong DependencyResolver

```go
dependency := NewDependency(model) // model là pointer hiện tại, ví dụ &data

for _, ref := range references {
    if ref.ID == "" {
        if ref.Required {
            return ErrReferenceRequired
        }
        continue
    }

    // Phase 1: resolver uses prefix to find/check the reference.
    message := registry.Resolve(ctx, ref.Name, model, prefix)
    if message.HasError() {
        return referenceError(ref, message)
    }

    if ref.Target != "" {
        // Phase 2: apply only uses the current model pointer and ref.Target.
        if err := dependency.Apply(ref, message.Data); err != nil {
            return err
        }
    }
}
```

### Hook và persistence

Không cần thay đổi flow repository hiện tại vì `BeforeInsert` đã nhận pointer tới data:

```text
Repository.Create(data)
    ↓
hooks.BeforeInsert(&data)
    ↓
DependencyResolver.Resolve(&data)
    ↓
Dependency.Apply(ref{Target: "Route"}, route)
    ↓
db.Create(&data)
```

`CreateMany` cũng phải giữ object `_data` đã hydrate trước khi append vào danh sách documents.

### Kế hoạch triển khai

1. Thêm `Target` vào `Reference` và parse option `into=<FieldName>`.
2. Viết test parser: tag hợp lệ, target rỗng, option không hợp lệ.
3. Dùng resolver interface `func(ctx context.Context, data any, prefix string) types.Message`; resolver ép kiểu `data` về model phù hợp để lấy ID và dùng `prefix` khi query/check.
4. Viết `Dependency`/`decodeAndSet` với test cho struct, pointer-to-struct, map-to-struct, field không tồn tại và sai kiểu.
5. Cập nhật `DependencyResolver`: tạo `Dependency` từ pointer model, resolve một lần, validate error, sau đó apply vào `Target`.
6. Cập nhật hook test: chạy `BeforeInsert` và kiểm tra field target đã có data trước DB call.
7. Quyết định từng model dùng embedded snapshot (`bson:"route"`) hay runtime-only field (`bson:"-"`).
8. Sau khi API ổn định, deprecate flow merge `Message.Data` vào toàn bộ model để tránh resolver ghi đè field không liên quan.

## 20. Testing

Cần test:

- Local Resolver
- gRPC Resolver
- HTTP Resolver
- Registry
- Extractor
- Validator
- Hook
- Required/Optional reference
- Not Found
- Timeout
- Service Unavailable
- Resolver Not Found

## 21. Acceptance Criteria

- Model khai báo được `ref:"route"`.
- Local reference dùng Repository.
- Remote reference hỗ trợ gRPC/HTTP.
- Có thể thay Resolver mà không sửa Model.
- Hook không tạo DB connection.
- Hook không tạo Repository.
- Repository chỉ quản lý persistence.
- Model không phụ thuộc transport.
- Không circular dependency.
- Cross-service resolver không truy cập DB của service khác.

## 22. Final Architecture

```text
                         Model
                    ref:"cruise"
                           │
                           ▼
                  Generic Reference Hook
                           │
                           ▼
                  DependencyResolver
                           │
                           ▼
                  Reference Registry
                           │
             ┌─────────────┼─────────────┐
             ▼             ▼             ▼
       LocalResolver  GRPCResolver  HTTPResolver
             │             │             │
             ▼             ▼             ▼
        Repository     gRPC Client   HTTP Client
             │             │             │
             ▼             ▼             ▼
          MongoDB      Remote Service  Remote Service
```

## 23. Core Principles

> **Reference declaration belongs to Model.**

> **Reference dependency resolution belongs to DependencyResolver.Resolve.**

> **Resolution strategy belongs to Resolver.**

> **Local persistence belongs to Repository.**

> **Remote communication belongs to Client.**

> **Transport must never be hard-coded into Model.**

Thiết kế này cho phép cùng một cơ chế `ref` hoạt động trong Monolith, Modular Monolith và Microservices mà không làm Model, Repository hoặc Hook phụ thuộc trực tiếp vào transport hay service cụ thể.
