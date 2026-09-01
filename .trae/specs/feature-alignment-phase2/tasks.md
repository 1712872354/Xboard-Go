# Tasks

- [x] Task 1: 集成支付网关管理路由
  - [x] SubTask 1.1: 在router.go中导入PaymentGatewayHandler
  - [x] SubTask 1.2: 在router.go中初始化PaymentRepository、PaymentGatewayService和PaymentGatewayHandler
  - [x] SubTask 1.3: 在admin路由组中添加payment/gateways路由（CRUD、状态、排序）
  - [x] SubTask 1.4: 添加公开的支付方式列表接口（/api/v1/payment/methods）
- [x] Task 2: 集成设备状态管理路由
  - [x] SubTask 2.1: 在router.go中初始化DeviceService
  - [x] SubTask 2.2: 创建DeviceHandler处理器
  - [x] SubTask 2.3: 在admin路由组中添加设备状态查询路由
- [x] Task 3: 完善用户管理增强功能
  - [x] SubTask 3.1: 在UserHandler中添加GenerateUsers处理器方法
  - [x] SubTask 3.2: 在UserHandler中添加ExportUsersCSV处理器方法
  - [x] SubTask 3.3: 在admin/users路由组中添加批量生成和导出CSV路由
  - [ ] SubTask 3.4: 添加发送邮件功能的处理器和路由（未实现）
- [x] Task 4: 完善节点管理增强功能
  - [x] SubTask 4.1: 在NodeHandler中添加CopyNode处理器方法
  - [x] SubTask 4.2: 在NodeHandler中添加ResetNodeTraffic处理器方法
  - [x] SubTask 4.3: 在NodeHandler中添加BatchResetNodeTraffic处理器方法
  - [x] SubTask 4.4: 在admin/nodes路由组中添加复制、重置流量路由
- [x] Task 5: 补充管理员工单回复路由
  - [x] SubTask 5.1: 确认admin/tickets路由组已包含reply和close路由
  - [x] SubTask 5.2: 验证工单回复功能正常工作
- [x] Task 6: 验证所有新增路由
  - [x] SubTask 6.1: 运行go build确保编译通过
  - [x] SubTask 6.2: 运行go test确保现有测试通过

# Task Dependencies

- Task 1 无依赖，可独立完成
- Task 2 无依赖，可独立完成
- Task 3 无依赖，可独立完成
- Task 4 无依赖，可独立完成
- Task 5 无依赖，可独立完成
- Task 6 依赖 Task 1-5 完成

