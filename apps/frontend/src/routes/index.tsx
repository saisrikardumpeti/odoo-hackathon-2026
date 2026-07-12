import { usePing } from '#/lib/hooks/usePing'
import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/')({ component: Home })

function Home() {
  const { data, isLoading, isError } = usePing()
  if (isLoading) {
    return <>Loading</>
  }
  if (isError) {
    return <>{isError}</>
  }
  return (
    <div className="p-8">
      <h1 className="text-4xl font-bold">Welcome to TanStack Start</h1>
      <p className="mt-4 text-lg">
        Edit <code>src/routes/index.tsx</code> to get started.
      </p>
      <>{JSON.stringify(data, null, 2)}</>
    </div>
  )
}
