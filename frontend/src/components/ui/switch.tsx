import * as RS from '@radix-ui/react-switch'

export function Switch(props: { checked: boolean; onCheckedChange: (v: boolean) => void; disabled?: boolean; 'aria-label': string }) {
  return (
    <RS.Root
      {...props}
      className="relative inline-flex h-6 w-11 shrink-0 items-center rounded-full bg-border transition-colors data-[state=checked]:bg-accent disabled:opacity-50"
    >
      <RS.Thumb className="block size-5 translate-x-0.5 rounded-full bg-white shadow transition-transform data-[state=checked]:translate-x-[22px]" />
    </RS.Root>
  )
}
