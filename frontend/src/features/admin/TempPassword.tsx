import { Check, Copy } from 'lucide-react'
import { useState } from 'react'
import { Button } from '@/components/ui/button'

/** Shown once after creating an account or resetting a password. */
export function TempPassword({ code, password }: { code: string; password: string }) {
  const [copied, setCopied] = useState(false)
  return (
    <div className="space-y-3">
      <p className="text-sm">
        Give these to the employee. They must choose a new password when they first sign in. <strong>This password is shown only once.</strong>
      </p>
      <dl className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-2 rounded-lg bg-bg p-4 text-sm">
        <dt className="text-muted">Employee code</dt>
        <dd className="font-mono">{code}</dd>
        <dt className="text-muted">Temporary password</dt>
        <dd className="font-mono text-base tracking-wider select-all">{password}</dd>
      </dl>
      <Button
        variant="secondary"
        className="w-full"
        onClick={async () => {
          await navigator.clipboard?.writeText(password)
          setCopied(true)
        }}
      >
        {copied ? <Check className="size-4" /> : <Copy className="size-4" />}
        {copied ? 'Copied' : 'Copy password'}
      </Button>
    </div>
  )
}
