import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Bell, ChevronRight } from 'lucide-react'
import api from '@/lib/api'
import { formatDate } from '@/lib/utils'
import type { Notice } from '@/types'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { Separator } from '@/components/ui/separator'
import {
  Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle,
} from '@/components/ui/dialog'

export default function NoticesPage() {
  const [selectedNotice, setSelectedNotice] = useState<Notice | null>(null)

  const { data: notices, isLoading } = useQuery({
    queryKey: ['user', 'notices'],
    queryFn: async () => (await api.get('/notices')) as unknown as Notice[],
  })

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">公告</h1>

      {/* Notice List */}
      {isLoading ? (
        <div className="space-y-4">
          {Array.from({ length: 5 }).map((_, i) => (
            <Skeleton key={i} className="h-20 w-full" />
          ))}
        </div>
      ) : notices?.length ? (
        <div className="space-y-3">
          {notices.map((notice) => (
            <Card
              key={notice.id}
              className="cursor-pointer transition-colors hover:bg-muted/50"
              onClick={() => setSelectedNotice(notice)}
            >
              <CardContent className="flex items-center gap-4 p-4">
                <Bell className="h-5 w-5 text-muted-foreground shrink-0" />
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <p className="font-medium truncate">{notice.title}</p>
                    {new Date(notice.created_at) > new Date(Date.now() - 7 * 24 * 60 * 60 * 1000) && (
                      <Badge variant="default" className="text-xs">新</Badge>
                    )}
                  </div>
                  <p className="text-sm text-muted-foreground line-clamp-1 mt-1">
                    {notice.content.replace(/<[^>]*>/g, '').slice(0, 100)}
                  </p>
                  <p className="text-xs text-muted-foreground mt-1">
                    {formatDate(notice.created_at)}
                  </p>
                </div>
                <ChevronRight className="h-4 w-4 text-muted-foreground shrink-0" />
              </CardContent>
            </Card>
          ))}
        </div>
      ) : (
        <Card>
          <CardContent className="py-8 text-center text-muted-foreground">
            暂无公告
          </CardContent>
        </Card>
      )}

      {/* Notice Detail Dialog */}
      <Dialog open={!!selectedNotice} onOpenChange={() => setSelectedNotice(null)}>
        <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>{selectedNotice?.title}</DialogTitle>
            <DialogDescription>
              {selectedNotice && formatDate(selectedNotice.created_at)}
            </DialogDescription>
          </DialogHeader>
          <Separator />
          <div className="prose prose-sm dark:prose-invert max-w-none">
            {selectedNotice?.img_url && (
              <img
                src={selectedNotice.img_url}
                alt={selectedNotice.title}
                className="rounded-lg w-full object-cover max-h-48 mb-4"
              />
            )}
            <div
              dangerouslySetInnerHTML={{ __html: selectedNotice?.content || '' }}
            />
          </div>
        </DialogContent>
      </Dialog>
    </div>
  )
}
