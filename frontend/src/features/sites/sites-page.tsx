import { useState } from 'react'
import { useQuery, useMutation } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { api, ApiError } from '@/api/client'
import { qk } from '@/api/queryKeys'
import { queryClient } from '@/app/query-client'
import type { Site } from '@/api/types'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input, Label, FieldError } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Modal } from '@/components/ui/modal'
import { Pagination } from '@/components/ui/pagination'
import { LoadingState, ErrorState, EmptyState, AccessDenied } from '@/components/feedback/states'
import { useToast } from '@/components/feedback/toast'

const siteSchema = z.object({
  name: z.string().min(1, 'Required'),
  address: z.string().min(1, 'Required'),
  latitude: z.coerce.number().min(-90).max(90),
  longitude: z.coerce.number().min(-180).max(180),
  allowed_radius_m: z.coerce.number().int().min(10, 'At least 10 m').max(5000),
})
type SiteFormInput = z.input<typeof siteSchema>
type SiteForm = z.output<typeof siteSchema>

export function SitesPage() {
  const [page, setPage] = useState(1)
  const [editing, setEditing] = useState<Site | 'new' | null>(null)
  const toast = useToast()

  const { data, isLoading, error, refetch } = useQuery({
    queryKey: qk.staffSites({ page }),
    queryFn: () => api.getPaged<Site>(`/staff/sites?page=${page}&page_size=20`),
  })

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-end">
        <Button onClick={() => setEditing('new')}>Add site</Button>
      </div>

      {isLoading && <LoadingState />}
      {error instanceof ApiError && error.status === 403 && <AccessDenied />}
      {error && !(error instanceof ApiError && error.status === 403) && (
        <ErrorState description={error.message} onRetry={() => refetch()} />
      )}
      {data && data.items.length === 0 && (
        <EmptyState title="No sites yet" description="Add the workplaces where trainees report." />
      )}
      {data && data.items.length > 0 && (
        <Card>
          <CardContent className="overflow-x-auto p-0">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-[var(--color-border)] text-left text-[var(--color-text-muted)]">
                  <th className="px-4 py-3 font-medium">Name</th>
                  <th className="px-4 py-3 font-medium">Address</th>
                  <th className="px-4 py-3 font-medium">Radius</th>
                  <th className="px-4 py-3 font-medium">Status</th>
                  <th className="px-4 py-3" />
                </tr>
              </thead>
              <tbody>
                {data.items.map((s) => (
                  <tr key={s.id} className="border-b border-[var(--color-border)] last:border-0">
                    <td className="px-4 py-3 font-medium">{s.name}</td>
                    <td className="px-4 py-3 text-[var(--color-text-muted)]">{s.address}</td>
                    <td className="px-4 py-3">{s.allowed_radius_m} m</td>
                    <td className="px-4 py-3">
                      <Badge variant={s.is_active ? 'success' : 'default'}>
                        {s.is_active ? 'Active' : 'Inactive'}
                      </Badge>
                    </td>
                    <td className="px-4 py-3 text-right">
                      <Button variant="ghost" size="sm" onClick={() => setEditing(s)}>
                        Edit
                      </Button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </CardContent>
        </Card>
      )}
      {data && <Pagination meta={data.meta} onPage={setPage} />}

      {editing && (
        <SiteModal
          site={editing === 'new' ? null : editing}
          onClose={() => setEditing(null)}
          onSaved={(msg) => {
            setEditing(null)
            toast.success(msg)
            void queryClient.invalidateQueries({ queryKey: ['staff', 'sites'] })
          }}
        />
      )}
    </div>
  )
}

function SiteModal({
  site,
  onClose,
  onSaved,
}: {
  site: Site | null
  onClose: () => void
  onSaved: (msg: string) => void
}) {
  const {
    register,
    handleSubmit,
    setError,
    formState: { errors },
  } = useForm<SiteFormInput, unknown, SiteForm>({
    resolver: zodResolver(siteSchema),
    defaultValues: site
      ? {
          name: site.name,
          address: site.address,
          latitude: site.latitude,
          longitude: site.longitude,
          allowed_radius_m: site.allowed_radius_m,
        }
      : { allowed_radius_m: 100 },
  })
  const [active, setActive] = useState(site?.is_active ?? true)

  const save = useMutation({
    mutationFn: async (f: SiteForm) => {
      if (site) {
        return api.patch<Site>(`/staff/sites/${site.id}`, { ...f, is_active: active })
      }
      return api.post<Site>('/staff/sites', f)
    },
    onSuccess: () => onSaved(site ? 'Site updated' : 'Site created'),
    onError: (e) => {
      if (e instanceof ApiError && Object.keys(e.fields).length > 0) {
        for (const [k, v] of Object.entries(e.fields)) {
          setError(k as keyof SiteFormInput, { message: v })
        }
      }
    },
  })

  return (
    <Modal open onClose={onClose} title={site ? 'Edit site' : 'Add site'}>
      <form onSubmit={handleSubmit((f) => save.mutate(f))} className="space-y-3">
        <div>
          <Label htmlFor="name">Site name</Label>
          <Input id="name" invalid={!!errors.name} {...register('name')} />
          <FieldError>{errors.name?.message}</FieldError>
        </div>
        <div>
          <Label htmlFor="address">Address</Label>
          <Input id="address" invalid={!!errors.address} {...register('address')} />
          <FieldError>{errors.address?.message}</FieldError>
        </div>
        <div className="grid grid-cols-2 gap-3">
          <div>
            <Label htmlFor="latitude">Latitude</Label>
            <Input id="latitude" type="number" step="any" invalid={!!errors.latitude} {...register('latitude')} />
            <FieldError>{errors.latitude?.message}</FieldError>
          </div>
          <div>
            <Label htmlFor="longitude">Longitude</Label>
            <Input id="longitude" type="number" step="any" invalid={!!errors.longitude} {...register('longitude')} />
            <FieldError>{errors.longitude?.message}</FieldError>
          </div>
        </div>
        <div>
          <Label htmlFor="radius">Allowed radius (meters)</Label>
          <Input id="radius" type="number" invalid={!!errors.allowed_radius_m} {...register('allowed_radius_m')} />
          <FieldError>{errors.allowed_radius_m?.message}</FieldError>
          <p className="mt-1 text-xs text-[var(--color-text-subtle)]">
            Trainees outside this radius are flagged, not blocked.
          </p>
        </div>
        {site && (
          <label className="flex items-center gap-2 text-sm">
            <input type="checkbox" checked={active} onChange={(e) => setActive(e.target.checked)} />
            Active
          </label>
        )}
        {save.isError && save.error instanceof ApiError && Object.keys(save.error.fields).length === 0 && (
          <p role="alert" className="text-sm text-[var(--color-danger)]">{save.error.message}</p>
        )}
        <div className="flex justify-end gap-2 pt-2">
          <Button type="button" variant="ghost" onClick={onClose}>Cancel</Button>
          <Button type="submit" loading={save.isPending}>Save</Button>
        </div>
      </form>
    </Modal>
  )
}
