import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import api from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { Skeleton } from '@/components/ui/skeleton'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog'
import { formatDate } from '@/lib/utils'
import { MessageSquare, X, RefreshCw, Loader2 } from 'lucide-react'

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

  const { data: ticket, isLoading } = useQuery({
    queryKey: ['admin', 'ticket', ticketId],
    queryFn: async () => await api.get(`/admin/tickets/${ticketId}`) as unknown as Ticket & { replies: TicketReply[] },
  })

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

      <div className="space-y-3 max-h-[400px] overflow-y-auto">
        {ticket?.replies?.map((reply) => (
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

  const { data, isLoading, refetch } = useQuery({
    queryKey: ['admin', 'tickets', statusFilter],
    queryFn: async () => {
      const params = statusFilter !== 'all' ? `?status=${statusFilter}` : ''
      return await api.get(`/admin/tickets${params}`) as unknown as { list: Ticket[]; total: number }
    },
  })

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
          <div className="flex gap-2">
            {['all', '0', '1', '2'].map((s) => (
              <Button
                key={s}
                variant={statusFilter === s ? 'default' : 'outline'}
                size="sm"
                onClick={() => setStatusFilter(s)}
              >
                {s === 'all' ? '全部' : statusMap[Number(s)]?.label}
              </Button>
            ))}
          </div>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-3">{Array.from({ length: 5 }).map((_, i) => <Skeleton key={i} className="h-12" />)}</div>
          ) : (
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
                  <TableHead>操作</TableHead>
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
                    <TableCell>
                      <Button variant="ghost" size="sm" onClick={() => setSelectedTicket(t.id)}>
                        <MessageSquare className="h-4 w-4" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
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
