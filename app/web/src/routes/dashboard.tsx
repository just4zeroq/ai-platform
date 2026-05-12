import { createFileRoute } from '@tanstack/react-router'
import { useEffect, useState } from 'react'
import { useAuthStore } from '../stores/auth'
import { apiGet } from '../api/client'

export const Route = createFileRoute('/dashboard')({
  component: DashboardPage,
})

interface Balance {
  balance: number
  total_recharged: number
  total_consumed: number
}

function DashboardPage() {
  const token = useAuthStore((s) => s.token)
  const user = useAuthStore((s) => s.user)
  const [balance, setBalance] = useState<Balance | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!token) return
    apiGet<{ balance: Balance }>('/asset/balance', token)
      .then((res) => setBalance(res.balance))
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [token])

  return (
    <div className="p-6 space-y-6">
      <h1 className="text-2xl font-bold">Dashboard</h1>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="border rounded-lg p-4">
          <p className="text-sm text-muted-foreground">Balance</p>
          <p className="text-2xl font-bold">
            {loading ? '...' : balance ? balance.balance.toFixed(2) : '0.00'}
          </p>
        </div>
        <div className="border rounded-lg p-4">
          <p className="text-sm text-muted-foreground">Total Recharged</p>
          <p className="text-2xl font-bold">
            {loading ? '...' : balance ? balance.total_recharged.toFixed(2) : '0.00'}
          </p>
        </div>
        <div className="border rounded-lg p-4">
          <p className="text-sm text-muted-foreground">Total Consumed</p>
          <p className="text-2xl font-bold">
            {loading ? '...' : balance ? balance.total_consumed.toFixed(2) : '0.00'}
          </p>
        </div>
      </div>

      <div className="border rounded-lg p-4">
        <h2 className="font-semibold mb-2">Account Info</h2>
        <p className="text-sm text-muted-foreground">Username: {user?.username}</p>
        <p className="text-sm text-muted-foreground">User ID: {user?.id}</p>
      </div>
    </div>
  )
}
