# 功能对齐第二阶段 Spec

## Why
当前Go版本已实现核心业务功能，但与原版PHP相比仍有部分管理功能缺失。本次对齐专注于将已创建但未集成的服务、处理器和路由进行完整对接，使管理员能够使用完整的管理功能。

## What Changes
- 将PaymentGatewayHandler集成到路由系统，提供支付网关CRUD管理
- 将DeviceService集成到路由系统，提供设备状态查询和管理
- 为UserHandler添加批量生成用户和导出CSV的处理器方法和路由
- 为NodeHandler添加复制节点、重置流量的处理器方法和路由
- 补充管理员发送邮件、重置用户密钥等用户管理增强功能

## Impact
- Affected specs: 用户管理、节点管理、支付管理、设备状态
- Affected code: router.go, user_handler, node_handler, payment_gateway_handler

## ADDED Requirements

### Requirement: 支付网关管理
系统 SHALL 提供支付网关的完整CRUD管理功能，包括创建、查看、更新、删除、启用/禁用、排序。

#### Scenario: 管理员创建支付网关
- **WHEN** 管理员提交支付网关信息（名称、支付方式、配置等）
- **THEN** 系统创建支付网关并返回创建结果

#### Scenario: 管理员查看支付网关列表
- **WHEN** 管理员请求支付网关列表
- **THEN** 系统返回所有支付网关列表，按排序值排列

### Requirement: 设备状态管理
系统 SHALL 提供设备在线状态查询功能，支持查看用户在线设备和节点在线设备。

#### Scenario: 查看用户在线设备
- **WHEN** 管理员请求指定用户的在线设备
- **THEN** 系统返回该用户的在线设备列表

### Requirement: 用户批量管理
系统 SHALL 提供批量生成用户和导出用户CSV的功能。

#### Scenario: 批量生成用户
- **WHEN** 管理员指定生成数量、前缀、密码等参数
- **THEN** 系统批量创建用户并返回创建结果

#### Scenario: 导出用户CSV
- **WHEN** 管理员请求导出用户数据
- **THEN** 系统返回CSV格式的用户数据

### Requirement: 节点管理增强
系统 SHALL 提供节点复制和重置流量的功能。

#### Scenario: 复制节点
- **WHEN** 管理员选择一个节点进行复制
- **THEN** 系统创建节点副本，默认状态为离线

#### Scenario: 重置节点流量
- **WHEN** 管理员选择节点重置流量
- **THEN** 系统重置指定节点的流量统计

## MODIFIED Requirements

### Requirement: 路由注册
所有新增的处理器方法必须在router.go中正确注册路由。

## REMOVED Requirements
无
