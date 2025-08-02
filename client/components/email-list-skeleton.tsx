import { Skeleton } from "@/components/ui/skeleton";

export function EmailListSkeleton() {
  return (
    <div className="w-96 border-r bg-card/50 flex flex-col h-full">
      {/* Search and controls skeleton */}
      <div className="p-4 border-b space-y-3 flex-shrink-0">
        <Skeleton className="h-10 w-full" />
        <div className="flex items-center justify-start">
          <Skeleton className="h-8 w-16" />
        </div>
      </div>

      {/* Email list skeleton */}
      <div className="flex-1 overflow-y-auto">
        <div className="divide-y">
          {Array.from({ length: 8 }).map((_, index) => (
            <div key={index} className="p-4 space-y-2">
              <div className="flex items-start justify-between mb-2">
                <div className="flex-1 min-w-0 space-y-2">
                  <div className="flex items-center gap-2 mb-1">
                    <Skeleton className="h-4 w-32" />
                  </div>
                  <Skeleton className="h-4 w-48" />
                  <Skeleton className="h-3 w-full" />
                  <Skeleton className="h-3 w-3/4" />
                </div>
                <div className="flex flex-col items-end gap-1 ml-2">
                  <Skeleton className="h-3 w-12" />
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}