import { createFileRoute } from '@tanstack/react-router'

import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ProcessTable } from '@/features/processes/process-table'
import { ResourceTable } from '@/features/resources/resource-table'

export const Route = createFileRoute('/setup')({
  component: SetupPage,
})

function SetupPage() {
  return (
    <div className="flex flex-col gap-6">
      <header>
        <h1 className="text-2xl font-semibold">Setup</h1>
        <p className="text-muted-foreground text-sm">
          Define the resources each step consumes, then the ordered steps
          themselves.
        </p>
      </header>

      <Tabs defaultValue="processes">
        <TabsList>
          <TabsTrigger value="processes">Processes</TabsTrigger>
          <TabsTrigger value="resources">Resources</TabsTrigger>
        </TabsList>
        <TabsContent value="processes" className="mt-4">
          <ProcessTable />
        </TabsContent>
        <TabsContent value="resources" className="mt-4">
          <ResourceTable />
        </TabsContent>
      </Tabs>
    </div>
  )
}
