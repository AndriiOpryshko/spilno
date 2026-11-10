import * as React from "react"

import { Input } from "@/components/ui/input"

// A native date input (YYYY-MM-DD) in the shadcn Input style.
function DateField(props: Omit<React.ComponentProps<"input">, "type">) {
  return <Input type="date" data-slot="date-field" {...props} />
}

export { DateField }
