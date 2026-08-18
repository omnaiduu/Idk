import { Skeleton } from "@/components/ui/skeleton";

export function BenchSkeleton() {
  return (
    <div className="flex h-screen w-full flex-col bg-void">
      <div className="flex h-12 items-center gap-4 border-b border-cream-border px-4">
        <Skeleton className="h-4 w-32" />
        <Skeleton className="h-4 w-20" />
        <Skeleton className="ml-auto h-7 w-16 rounded-full" />
        <Skeleton className="h-8 w-14" />
        <Skeleton className="h-8 w-14" />
      </div>
      <div className="flex min-h-0 flex-1">
        <div className="w-[200px] space-y-2 border-r border-cream-border p-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <Skeleton key={i} className="h-14 w-full rounded-md" />
          ))}
        </div>
        <div className="min-w-0 flex-1 p-3">
          <Skeleton className="h-full w-full rounded-md" />
        </div>
        <div className="w-[280px] space-y-4 border-l border-cream-border p-2">
          <Skeleton className="mx-auto h-64 w-64 rounded-sm" />
          <Skeleton className="h-8 w-full" />
        </div>
      </div>
      <div className="h-[180px] border-t border-cream-border p-3">
        <Skeleton className="h-full w-full rounded-md" />
      </div>
    </div>
  );
}
