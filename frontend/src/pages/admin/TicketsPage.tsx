import { useState, useCallback } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import api from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { formatDate, formatCurrency } from '@/lib/utils'
import { MessageSquare, X, RefreshCw, Loader2, Search, User as UserIcon, CreditCard, Calendar, Package } from 'lucide-react'
import { usePlans } from '@/hooks/usePlans'

interface Ticket {
  id: number
  user_id: number
  subject: string
  category: number
  priority: number
  status: number
  user?: { email: string }
  created_at: string
  updated_at: string
}

interface TicketDetailUser {
  id: number
  email: string
  balance: number
  commission: number
  plan_id?: number | null
  expired_at?: string | null
  traffic_limit: number
  used_traffic: number
}

interface TicketReply {
  id: number
  content: string
  is_admin: boolean
  user?: { email: string }
  created_at: string
}

const statusMap: Record<number, { label: string; variant: 'default' | 'secondary' | 'success' | 'destructive' | 'warning' }> = {
  0: { label: '待处理', variant: 'warning' },
  1: { label: '已回复', variant: 'default' },
  2: { label: '已关闭', variant: 'secondary' },
  3: { label: '处理中', variant: 'default' },
}

const priorityMap: Record<number, string> = {
  0: '低',
  1: '普通',
  2: '高',
}

const categoryMap: Record<number, string> = {
  0: '一般',
  1: '账单',
  2: '技术',
  3: '账户',
}

function TicketDetail({ ticketId, onClose }: { ticketId: number; onClose: () => void }) {
  const qc = useQueryClient()
  const [replyContent, setReplyContent] = useState('')
  const { data: plans } = usePlans()

  const { data: ticketData, isLoading } = useQuery({
    queryKey: ['admin', 'ticket', ticketId],
    queryFn: async () => {
      const res = await api.get(`/admin/tickets/${ticketId}`) as unknown as { ticket: Ticket & { user?: TicketDetailUser }; replies: TicketReply[] }
      return res
    },
  })

  const ticket = ticketData?.ticket
  const replies = ticketData?.replies

  const { data: userData } = useQuery({
    queryKey: ['admin', 'user', ticket?.user_id],
    queryFn: async () => await api.get(`/admin/users/${ticket?.user_id}`) as unknown as TicketDetailUser,
    enabled: !!ticket?.user_id,
  })

  const userInfo = userData || ticket?.user

  const replyMutation = useMutation({
    mutationFn: async (content: string) => await api.post(`/admin/tickets/${ticketId}/reply`, { content }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'ticket', ticketId] })
      qc.invalidateQueries({ queryKey: ['admin', 'tickets'] })
      setReplyContent('')
      toast.success('回复成功')
    },
    onError: () => toast.error('回复失败'),
  })

  const closeMutation = useMutation({
    mutationFn: async () => await api.post(`/admin/tickets/${ticketId}/close`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'ticket', ticketId] })
      qc.invalidateQueries({ queryKey: ['admin', 'tickets'] })
      toast.success('工单已关闭')
    },
    onError: () => toast.error('关闭失败'),
  })

  if (isLoading) return <div className="space-y-3">{Array.from({ length: 3 }).map((_, i) => <Skeleton key={i} className="h-20" />)}</div>

  const planName = userInfo?.plan_id ? (plans?.find((p) => p.id === userInfo.plan_id)?.name || `#${userInfo.plan_id}`) : '无套餐'

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-semibold">{ticket?.subject}</h3>
          <p className="text-sm text-muted-foreground">
            {ticket?.user?.email} · {categoryMap[ticket?.category ?? 0]} · {priorityMap[ticket?.priority ?? 1]}
          </p>
        </div>
        <Badge variant={statusMap[ticket?.status ?? 0]?.variant}>
          {statusMap[ticket?.status ?? 0]?.label}
        </Badge>
      </div>

      {userInfo && (
        <Card className="bg-muted/30">
          <CardContent className="p-4">
            <div className="flex items-center gap-2 mb-3">
              <UserIcon className="h-4 w-4 text-muted-foreground" />
              <span className="text-sm font-medium">用户信息</span>
            </div>
            <div className="grid grid-cols-2 gap-3 text-sm">
              <div className="flex items-center gap-2">
                <span className="text-muted-foreground">邮箱：</span>
                <span className="font-medium">{userInfo.email || ticket?.user?.email || '-'}</span>
              </div>
              <div className="flex items-center gap-2">
                <CreditCard className="h-3.5 w-3.5 text-muted-foreground" />
                <span className="text-muted-foreground">余额：</span>
                <span className="font-medium">{formatCurrency(userInfo.balance ?? 0)}</span>
              </div>
              <div className="flex items-center gap-2">
                <CreditCard className="h-3.5 w-3.5 text-muted-foreground" />
                <span className="text-muted-foreground">佣金余额：</span>
                <span className="font-medium">{formatCurrency(userInfo.commission ?? 0)}</span>
              </div>
              <div className="flex items-center gap-2">
                <Package className="h-3.5 w-3.5 text-muted-foreground" />
                <span className="text-muted-foreground">套餐：</span>
                <span className="font-medium">{planName}</span>
              </div>
              <div className="flex items-center gap-2 col-span-2">
                <Calendar className="h-3.5 w-3.5 text-muted-foreground" />
                <span className="text-muted-foreground">到期时间：</span>
                <span className="font-medium">{userInfo.expired_at ? formatDate(userInfo.expired_at) : '永不过期'}</span>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      <div className="space-y-3 max-h-[400px] overflow-y-auto">
        {replies?.map((reply) => (
          <div key={reply.id} className={`rounded-lg p-3 ${reply.is_admin ? 'bg-primary/5 ml-8' : 'bg-muted mr-8'}`}>
            <div className="flex items-center gap-2 mb-1">
              <span className="text-sm font-medium">{reply.is_admin ? '管理员' : reply.user?.email}</span>
              <span className="text-xs text-muted-foreground">{formatDate(reply.created_at)}</span>
            </div>
            <p className="text-sm whitespace-pre-wrap">{reply.content}</p>
          </div>
        ))}
      </div>

      {ticket?.status !== 2 && (
        <div className="space-y-2">
          <Textarea
            placeholder="输入回复内容..."
            value={replyContent}
            onChange={(e) => setReplyContent(e.target.value)}
            rows={3}
          />
          <div className="flex gap-2">
            <Button
              onClick={() => replyMutation.mutate(replyContent)}
              disabled={!replyContent.trim() || replyMutation.isPending}
            >
              {replyMutation.isPending ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <MessageSquare className="mr-2 h-4 w-4" />}
              回复
            </Button>
            <Button variant="outline" onClick={() => closeMutation.mutate()} disabled={closeMutation.isPending}>
              <X className="mr-2 h-4 w-4" />
              关闭工单
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}

export default function TicketsPage() {
  const [selectedTicket, setSelectedTicket] = useState<number | null>(null)
  const [statusFilter, setStatusFilter] = useState<string>('all')
  const [replyStatus, setReplyStatus] = useState<string>('all')
  const [keyword, setKeyword] = useState('')
  const [searchKeyword, setSearchKeyword] = useState('')
  const [page, setPage] = useState(1)
  const pageSize = 20

  const { data, isLoading, refetch } = useQuery({
    queryKey: ['admin', 'tickets', statusFilter, replyStatus, searchKeyword, page],
    queryFn: async () => {
      const params = new URLSearchParams()
      params.set('page', String(page))
      params.set('page_size', String(pageSize))
      if (statusFilter !== 'all') params.set('status', statusFilter)
      if (replyStatus !== 'all') params.set('reply_status', replyStatus)
      if (searchKeyword) params.set('keyword', searchKeyword)
      return await api.get(`/admin/tickets?${params.toString()}`) as unknown as { list: Ticket[]; total: number }
    },
  })

  const handleSearch = useCallback(() => {
    setSearchKeyword(keyword.trim())
    setPage(1)
  }, [keyword])

  const handleResetSearch = useCallback(() => {
    setKeyword('')
    setSearchKeyword('')
  }, [])

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">工单管理</h1>
          <p className="text-muted-foreground">管理用户提交的工单</p>
        </div>
        <Button variant="outline" size="sm" onClick={() => refetch()}>
          <RefreshCw className="mr-2 h-4 w-4" />
          刷新
        </Button>
      </div>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle>工单列表</CardTitle>
          <div className="flex gap-2 items-center">
            {['all', '0', '1', '3', '2'].map((s) => (
              <Button
                key={s}
                variant={statusFilter === s ? 'default' : 'outline'}
                size="sm"
                onClick={() => { setStatusFilter(s); setPage(1) }}
              >
                {s === 'all' ? '全部' : statusMap[Number(s)]?.label}
              </Button>
            ))}
            <Select value={replyStatus} onValueChange={(v) => { setReplyStatus(v); setPage(1) }}>
              <SelectTrigger className="h-8 w-[120px]">
                <SelectValue placeholder="回复状态" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部回复状态</SelectItem>
                <SelectItem value="0">待回复</SelectItem>
                <SelectItem value="1">已回复</SelectItem>
              </SelectContent>
            </Select>
            <div className="flex items-center gap-1 ml-2">
              <Input
                placeholder="搜索用户邮箱..."
                value={keyword}
                onChange={(e) => setKeyword(e.target.value)}
                onKeyDown={(e) => { if (e.key === 'Enter') handleSearch() }}
                className="h-8 w-[200px]"
              />
              <Button variant="outline" size="sm" className="h-8 px-2" onClick={handleSearch}>
                <Search className="h-4 w-4" />
              </Button>
              {searchKeyword && (
                <Button variant="ghost" size="sm" className="h-8 px-2" onClick={handleResetSearch}>
                  <X className="h-4 w-4" />
                </Button>
              )}
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-3">{Array.from({ length: 5 }).map((_, i) => <Skeleton key={i} className="h-12" />)}</div>
          ) : (
            <>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>ID</TableHead>
                  <TableHead>主题</TableHead>
                  <TableHead>用户</TableHead>
                  <TableHead>分类</TableHead>
                  <TableHead>优先级</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>更新时间</TableHead>
                  <TableHead className="text-right">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data?.list?.length === 0 ? (
                  <TableRow><TableCell colSpan={8} className="h-24 text-center text-muted-foreground">暂无工单</TableCell></TableRow>
                ) : data?.list?.map((t) => (
                  <TableRow key={t.id}>
                    <TableCell className="font-mono text-xs">#{t.id}</TableCell>
                    <TableCell className="font-medium max-w-[200px] truncate">{t.subject}</TableCell>
                    <TableCell className="text-sm">{t.user?.email}</TableCell>
                    <TableCell><Badge variant="outline">{categoryMap[t.category]}</Badge></TableCell>
                    <TableCell><Badge variant={t.priority === 2 ? 'destructive' : t.priority === 1 ? 'default' : 'secondary'}>{priorityMap[t.priority]}</Badge></TableCell>
                    <TableCell><Badge variant={statusMap[t.status]?.variant}>{statusMap[t.status]?.label}</Badge></TableCell>
                    <TableCell className="text-xs text-muted-foreground">{formatDate(t.updated_at)}</TableCell>
                    <TableCell className="text-right">
                      <Button variant="ghost" size="sm" onClick={() => setSelectedTicket(t.id)}>
                        <MessageSquare className="h-4 w-4" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
            {data && data.total > pageSize && (
              <div className="mt-4 flex items-center justify-between">
                <p className="text-sm text-muted-foreground">共 {data.total} 条</p>
                <div className="flex gap-2">
                  <Button size="sm" variant="outline" disabled={page <= 1} onClick={() => setPage(page - 1)}>上一页</Button>
                  <span className="flex items-center text-sm">{page} / {Math.ceil(data.total / pageSize)}</span>
                  <Button size="sm" variant="outline" disabled={page >= Math.ceil(data.total / pageSize)} onClick={() => setPage(page + 1)}>下一页</Button>
                </div>
              </div>
            )}
            </>
          )}
        </CardContent>
      </Card>

      <Dialog open={selectedTicket !== null} onOpenChange={() => setSelectedTicket(null)}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>工单详情</DialogTitle>
            <DialogDescription>查看和回复工单</DialogDescription>
          </DialogHeader>
          {selectedTicket && <TicketDetail ticketId={selectedTicket} onClose={() => setSelectedTicket(null)} />}
        </DialogContent>
      </Dialog>
    </div>
  )
}
