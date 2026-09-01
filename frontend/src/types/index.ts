export type UserRole = 'user' | 'admin'
export type UserStatus = 0 | 1
export type NodeStatus = 0 | 1 | 2
export type OrderStatus = 0 | 1 | 2 | 3
export type TicketStatus = 0 | 1 | 2

export interface User {
  id: number
  uuid?: string
  email: string
  role: UserRole
  status: UserStatus
  traffic_limit: number
  used_traffic: number
  expired_at: string | null
  subscribe_token: string
  two_factor_enabled: boolean
  balance: number
  commission: number
  commission_type?: number
  commission_rate?: number
  plan_id?: number | null
  group_id?: number
  discount?: number
  speed_limit?: number
  device_limit?: number
  online_count: number
  remind_expire: boolean
  remind_traffic: boolean
  invite_code_id?: number
  inviter_id?: number
  remarks?: string
  phone?: string
  last_login_at?: string
  last_login_ip?: string
  is_staff?: boolean
  created_at: string
  updated_at: string
}

export interface LoginResponse {
  access_token: string
  refresh_token: string
  expires_in: number
  user: User
}

export interface Plan {
  id: number
  name: string
  price: number
  traffic: number
  duration_days: number
  device_limit: number
  speed_limit?: number
  node_group: string
  description: string
  content?: string
  prices?: string
  status: number
  show: number
  sell: number
  renew: number
  capacity_limit: number
  tags: string
  reset_traffic_method: number
}

export interface Node {
  id: number
  name: string
  type: string
  host: string
  port: number
  server_port: number
  code: string
  server_info: string
  group_ids: string
  rate: number
  status: NodeStatus
  parent_id: number
  machine_id: number | null
  enabled: boolean
  sort: number
  show: number
  tags: string
  online_user_count: number
  u: number          // upload traffic
  d: number          // download traffic
  transfer_enable: number  // traffic limit, 0 = unlimited
  rate_time_enable: boolean
  rate_time_ranges: string
  custom_outbounds: string
  custom_routes: string
  cert_config: string
  ports: string
  health_check_port: number
  health_check_interval: number
  health_check_timeout: number
  health_check_type: string
  last_online: string | null
  created_at: string
  updated_at: string
}

export interface Order {
  id: number
  trade_no: string
  user_id: number
  plan_id: number
  plan_name?: string
  amount: number
  coupon_code?: string
  discount?: number
  actual_amount?: number
  status: OrderStatus
  payment_method: string
  paid_at?: string
  created_at: string
  plan?: Plan
}

export interface Ticket {
  id: number
  user_id: number
  subject: string
  category: string
  priority: number
  status: TicketStatus
  reply_status?: number
  last_reply_user_id?: number
  created_at: string
  updated_at: string
  replies?: TicketReply[]
}

export interface TicketReply {
  id: number
  ticket_id: number
  user_id: number
  content: string
  is_admin: boolean
  created_at: string
}

export interface Notice {
  id: number
  title: string
  content: string
  img_url: string
  tags: string
  show: number
  sort: number
  groups: string
  popup?: boolean
  created_at: string
}

export interface Knowledge {
  id: number
  category: string
  title: string
  content: string
  language: string
  show: number
  sort: number
  created_at: string
}

export interface Coupon {
  id: number
  code: string
  name: string
  type: number
  value: number
  min_amount: number
  max_discount: number
  plan_ids: string
  user_ids: string
  limit_count: number
  used_count: number
  limit_period: string
  limit_use_with_user: number
  start_date: string
  end_date: string
  status: number
  created_at: string
}

export interface GiftCardTemplate {
  id: number
  name: string
  description: string
  type: number
  value: number
  traffic: number
  duration: number
  plan_id: number
  price: number
  status: number
  created_at: string
}

export interface GiftCardCode {
  id: number
  template_id: number
  code: string
  status: number
  user_id: number
  used_at: string
  created_at: string
}

export interface InviteCode {
  id: number
  user_id: number
  code: string
  commission: number
  used_count: number
  limit_count: number
  status: number
  created_at: string
}

export interface CommissionLog {
  id: number
  user_id: number
  from_user_id: number
  from_user_email?: string
  order_id: number
  amount: number
  order_amount: number
  commission: number
  status: number
  created_at: string
}

export interface MailTemplate {
  id: number
  name: string
  subject: string
  body: string
  language: string
  remark: string
  created_at: string
}

export interface Plugin {
  id: number
  name: string
  title: string
  description: string
  version: string
  author: string
  homepage: string
  config: string
  type: string
  status: number
  created_at: string
}

export interface AuditLog {
  id: number
  user_id: number
  username: string
  action: string
  resource: string
  resource_id: string
  detail: string
  ip: string
  user_agent: string
  created_at: string
}

export interface ServerMachine {
  id: number
  name: string
  remark?: string
  notes?: string
  host?: string
  port?: number
  protocol?: string
  status: number
  is_active?: boolean
  cpu: number
  memory: number
  disk: number
  uptime?: number
  token?: string
  nodes_count?: number
  node_count?: number
  last_check_at?: string
  last_seen_at?: number
  load_status?: string
  created_at: string
  updated_at?: string
}

export interface Setting {
  id: number
  key: string
  value: string
  group: string
  remark: string
}

export interface TrafficStats {
  total_used: number
  traffic_limit: number
  remaining: number
  today_upload: number
  today_download: number
  week_upload: number
  week_download: number
  month_upload: number
  month_download: number
}

export interface TrafficLog {
  id: number
  user_id: number
  node_id: number
  node_name?: string
  upload: number
  download: number
  recorded_at: string
}

export interface TrafficHistoryItem {
  date: string
  upload: number
  download: number
}

export interface DashboardOverview {
  total_users: number
  active_users: number
  total_orders: number
  paid_orders: number
  total_income: number
  today_income: number
  total_nodes: number
  online_nodes: number
  open_tickets: number
  unused_redeems: number
}

export interface DailyIncome {
  date: string
  amount: number
  count: number
}

export interface DailyUserGrowth {
  date: string
  new_users: number
}

export interface NodeStats {
  total: number
  online: number
  offline: number
  maintenance: number
}

export interface PaginatedResponse<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}

export interface RedeemCode {
  id: number
  code: string
  plan_id: number
  plan_name?: string
  status: number
  used_by: number
  used_at: string
  created_at: string
}

export interface RedeemStats {
  total: number
  used: number
  unused: number
}

export interface ServerGroup {
  id: number
  name: string
  description: string
  plan_ids: string
  sort: number
  status: number
  users_count?: number
  server_count?: number
  created_at: string
}

export interface PaymentMethod {
  id: string
  name: string
  icon?: string
  enabled: boolean
}

export interface PublicSettings {
  app_name: string
  app_url: string
  app_description: string
  app_logo: string
  tos_url: string
  frontend_theme: string
}

export interface ServerRoute {
  id: number
  group_id: number
  name: string
  match: string
  action: string
  action_value?: string
  sort: number
  status: number
  created_at: string
}
