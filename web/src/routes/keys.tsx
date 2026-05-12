import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/keys')({
  component: KeysPage,
})

function KeysPage() {
  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold">API Keys</h1>
    </div>
  )
}
