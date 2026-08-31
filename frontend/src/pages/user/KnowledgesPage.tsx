import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Search, BookOpen, ChevronRight } from 'lucide-react'
import api from '@/lib/api'
import { formatDate } from '@/lib/utils'
import type { Knowledge } from '@/types'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { Separator } from '@/components/ui/separator'

export default function KnowledgesPage() {
  const [search, setSearch] = useState('')
  const [selectedCategory, setSelectedCategory] = useState<string | null>(null)
  const [expandedId, setExpandedId] = useState<number | null>(null)

  const { data: knowledges, isLoading } = useQuery({
    queryKey: ['user', 'knowledges'],
    queryFn: async () => (await api.get('/knowledges')) as unknown as Knowledge[],
  })

  const categories = knowledges
    ? [...new Set(knowledges.map((k) => k.category))]
    : []

  const filtered = knowledges?.filter((k) => {
    const matchCategory = !selectedCategory || k.category === selectedCategory
    const matchSearch =
      !search ||
      k.title.toLowerCase().includes(search.toLowerCase()) ||
      k.content.toLowerCase().includes(search.toLowerCase())
    return matchCategory && matchSearch
  })

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">知识库</h1>

      {/* Search & Filter */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="搜索文章..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-9"
          />
        </div>
        <div className="flex flex-wrap gap-2">
          <Button
            variant={selectedCategory === null ? 'default' : 'outline'}
            size="sm"
            onClick={() => setSelectedCategory(null)}
          >
            全部
          </Button>
          {categories.map((cat) => (
            <Button
              key={cat}
              variant={selectedCategory === cat ? 'default' : 'outline'}
              size="sm"
              onClick={() => setSelectedCategory(cat)}
            >
              {cat}
            </Button>
          ))}
        </div>
      </div>

      {/* Knowledge List */}
      {isLoading ? (
        <div className="space-y-4">
          {Array.from({ length: 5 }).map((_, i) => (
            <Skeleton key={i} className="h-24 w-full" />
          ))}
        </div>
      ) : filtered?.length ? (
        <div className="space-y-4">
          {filtered.map((knowledge) => (
            <Card key={knowledge.id}>
              <CardContent className="p-0">
                <button
                  className="w-full text-left p-4 hover:bg-muted/50 transition-colors"
                  onClick={() => setExpandedId(expandedId === knowledge.id ? null : knowledge.id)}
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3 min-w-0">
                      <BookOpen className="h-5 w-5 text-muted-foreground shrink-0" />
                      <div className="min-w-0">
                        <p className="font-medium">{knowledge.title}</p>
                        <div className="flex items-center gap-2 mt-1">
                          <Badge variant="outline" className="text-xs">
                            {knowledge.category}
                          </Badge>
                          <span className="text-xs text-muted-foreground">
                            {formatDate(knowledge.created_at, { year: 'numeric', month: '2-digit', day: '2-digit' })}
                          </span>
                        </div>
                      </div>
                    </div>
                    <ChevronRight
                      className={`h-4 w-4 text-muted-foreground transition-transform ${
                        expandedId === knowledge.id ? 'rotate-90' : ''
                      }`}
                    />
                  </div>
                  {expandedId !== knowledge.id && (
                    <p className="text-sm text-muted-foreground mt-2 line-clamp-2">
                      {knowledge.content.replace(/<[^>]*>/g, '').slice(0, 150)}
                    </p>
                  )}
                </button>
                {expandedId === knowledge.id && (
                  <>
                    <Separator />
                    <div className="p-4 prose prose-sm dark:prose-invert max-w-none">
                      <div
                        dangerouslySetInnerHTML={{ __html: knowledge.content }}
                      />
                    </div>
                  </>
                )}
              </CardContent>
            </Card>
          ))}
        </div>
      ) : (
        <Card>
          <CardContent className="py-8 text-center text-muted-foreground">
            {search || selectedCategory ? '没有找到匹配的文章' : '暂无知识库文章'}
          </CardContent>
        </Card>
      )}
    </div>
  )
}
