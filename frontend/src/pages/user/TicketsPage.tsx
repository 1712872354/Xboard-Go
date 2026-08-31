import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import { Plus, Send, Loader2, MessageSquare, XCircle } from 'lucide-react'
import api from '@/lib/api'
import { formatDate } from '@/lib/utils'
import type { Ticket, TicketReply, PaginatedResponse } from '@/types'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select'
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from '@/components/ui/dialog'
import { Separator } from '@/components/ui/separator'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'

const statusMap: Record<number, { label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline' }> = {
  0: { label: '待回复', variant: 'outline' },
  1: { label: '已回复', variant: 'default' },
  2: { label: '已关闭', variant: 'secondary' },
}

const createSchema = z.object({
  subject: z.string().min(1, '请输入主题'),
  content: z.string().min(1, '请输入内容'),
  category: z.string().min(1, '请选择分类'),
  priority: z.string().min(1, '请选择优先级'),
})

type CreateFormData = z.infer<typeof createSchema>

export default function TicketsPage() {
  const queryClient = useQueryClient()
  const [showCreate, setShowCreate] = useState(false)
  const [selectedTicket, setSelectedTicket] = useState<Ticket | null>(null)
  const [replyContent, setReplyContent] = useState('')

  const { data: tickets, isLoading } = useQuery({
    queryKey: ['user', 'tickets'],
    queryFn: async () => (await api.get('/tickets')) as unknown as PaginatedResponse<Ticket>,
  })

  const { data: ticketDetail, isLoading: detailLoading } = useQuery({
    queryKey: ['user', 'tickets', selectedTicket?.id],
    queryFn: async () =>
      (await api.get(`/tickets/${selectedTicket?.id}`)) as unknown as { ticket: Ticket; replies: TicketReply[] },
    enabled: !!selectedTicket?.id,
  })

  const createForm = useForm<CreateFormData>({
    resolver: zodResolver(createSchema),
    defaultValues: { subject: '', content: '', category: '0', priority: '0' },
  })

  const createTicket = useMutation({
    mutationFn: async (data: CreateFormData) =>
      await api.post('/tickets', { ...data, category: Number(data.category), priority: Number(data.priority) }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['user', 'tickets'] })
      toast.success('工单创建成功')
      setShowCreate(false)
      createForm.reset()
    },
    onError: (err: any) => toast.error(err?.message || '创建失败'),
  })

  const replyTicket = useMutation({
    mutationFn: async () =>
      await api.post(`/tickets/${selectedTicket?.id}/reply`, { content: replyContent }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['user', 'tickets', selectedTicket?.id] })
      toast.success('回复成功')
      setReplyContent('')
    },
    onError: (err: any) => toast.error(err?.message || '回复失败'),
  })

  const closeTicket = useMutation({
    mutationFn: async () => await api.post(`/tickets/${selectedTicket?.id}/close`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['user', 'tickets'] })
      toast.success('工单已关闭')
      setSelectedTicket(null)
    },
    onError: (err: any) => toast.error(err?.message || '关闭失败'),
  })

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">工单</h1>
        <Button onClick={() => setShowCreate(true)}>
          <Plus className="mr-2 h-4 w-4" />
          创建工单
        </Button>
      </div>

      {/* Ticket List */}
      <div className="space-y-3">
        {isLoading ? (
          Array.from({ length: 5 }).map((_, i) => (
            <Skeleton key={i} className="h-20 w-full" />
          ))
        ) : tickets?.list?.length ? (
          tickets.list.map((ticket) => (
            <Card
              key={ticket.id}
              className="cursor-pointer transition-colors hover:bg-muted/50"
              onClick={() => setSelectedTicket(ticket)}
            >
              <CardContent className="flex items-center justify-between p-4">
                <div className="flex items-center gap-3 min-w-0">
                  <MessageSquare className="h-5 w-5 text-muted-foreground shrink-0" />
                  <div className="min-w-0">
                    <p className="font-medium truncate">{ticket.subject}</p>
                    <p className="text-sm text-muted-foreground">
                      {ticket.category} · {formatDate(ticket.created_at)}
                    </p>
                  </div>
                </div>
                <Badge variant={statusMap[ticket.status]?.variant || 'secondary'}>
                  {statusMap[ticket.status]?.label || '未知'}
                </Badge>
              </CardContent>
            </Card>
          ))
        ) : (
          <Card>
            <CardContent className="py-8 text-center text-muted-foreground">
              暂无工单
            </CardContent>
          </Card>
        )}
      </div>

      {/* Ticket Detail Dialog */}
      <Dialog open={!!selectedTicket} onOpenChange={() => setSelectedTicket(null)}>
        <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>{selectedTicket?.subject}</DialogTitle>
            <DialogDescription>
              {selectedTicket?.category} · {selectedTicket && formatDate(selectedTicket.created_at)}
            </DialogDescription>
          </DialogHeader>

          {detailLoading ? (
            <div className="space-y-4">
              {Array.from({ length: 3 }).map((_, i) => (
                <Skeleton key={i} className="h-20 w-full" />
              ))}
            </div>
          ) : (
            <div className="space-y-4">
              {ticketDetail?.replies?.map((reply: TicketReply) => (
                <div
                  key={reply.id}
                  className={`flex gap-3 ${reply.is_admin ? '' : 'flex-row-reverse'}`}
                >
                  <Avatar className="h-8 w-8 shrink-0">
                    <AvatarFallback>
                      {reply.is_admin ? 'A' : 'U'}
                    </AvatarFallback>
                  </Avatar>
                  <div
                    className={`rounded-lg p-3 max-w-[80%] ${
                      reply.is_admin
                        ? 'bg-muted'
                        : 'bg-primary text-primary-foreground ml-auto'
                    }`}
                  >
                    <p className="text-sm whitespace-pre-wrap">{reply.content}</p>
                    <p className={`text-xs mt-1 ${reply.is_admin ? 'text-muted-foreground' : 'text-primary-foreground/70'}`}>
                      {formatDate(reply.created_at)}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          )}

          <Separator />

          {selectedTicket?.status !== 2 && (
            <div className="space-y-3">
              <Textarea
                placeholder="输入回复内容..."
                value={replyContent}
                onChange={(e) => setReplyContent(e.target.value)}
                rows={3}
              />
              <div className="flex justify-between">
                <Button
                  variant="outline"
                  onClick={() => closeTicket.mutate()}
                  disabled={closeTicket.isPending}
                >
                  {closeTicket.isPending ? (
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  ) : (
                    <XCircle className="mr-2 h-4 w-4" />
                  )}
                  关闭工单
                </Button>
                <Button
                  onClick={() => replyTicket.mutate()}
                  disabled={!replyContent.trim() || replyTicket.isPending}
                >
                  {replyTicket.isPending ? (
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  ) : (
                    <Send className="mr-2 h-4 w-4" />
                  )}
                  发送
                </Button>
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>

      {/* Create Ticket Dialog */}
      <Dialog open={showCreate} onOpenChange={setShowCreate}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>创建工单</DialogTitle>
            <DialogDescription>描述您遇到的问题</DialogDescription>
          </DialogHeader>
          <form onSubmit={createForm.handleSubmit((d) => createTicket.mutate(d))}>
            <div className="space-y-4">
              <div className="space-y-2">
                <Label>主题</Label>
                <Input {...createForm.register('subject')} placeholder="简短描述您的问题" />
                {createForm.formState.errors.subject && (
                  <p className="text-sm text-destructive">
                    {createForm.formState.errors.subject.message}
                  </p>
                )}
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label>分类</Label>
                  <Select
                    value={createForm.watch('category')}
                    onValueChange={(v) => createForm.setValue('category', v)}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="选择分类" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="0">一般问题</SelectItem>
                      <SelectItem value="1">账单问题</SelectItem>
                      <SelectItem value="2">技术支持</SelectItem>
                      <SelectItem value="3">账户问题</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label>优先级</Label>
                  <Select
                    value={createForm.watch('priority')}
                    onValueChange={(v) => createForm.setValue('priority', v)}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="选择优先级" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="0">低</SelectItem>
                      <SelectItem value="1">中</SelectItem>
                      <SelectItem value="2">高</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>
              <div className="space-y-2">
                <Label>内容</Label>
                <Textarea
                  {...createForm.register('content')}
                  placeholder="详细描述您的问题..."
                  rows={5}
                />
                {createForm.formState.errors.content && (
                  <p className="text-sm text-destructive">
                    {createForm.formState.errors.content.message}
                  </p>
                )}
              </div>
            </div>
            <DialogFooter className="mt-4">
              <Button type="button" variant="outline" onClick={() => setShowCreate(false)}>
                取消
              </Button>
              <Button type="submit" disabled={createTicket.isPending}>
                {createTicket.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                提交
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  )
}
