import * as React from "react"

import { Input } from "@/components/ui/input"

// A native file input in the shadcn Input style.
function FileField(props: Omit<React.ComponentProps<"input">, "type">) {
  return <Input type="file" data-slot="file-field" {...props} />
}

export { FileField }
