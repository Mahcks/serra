"use client"

import { type ColumnDef } from "@tanstack/react-table"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import {
  CheckCircle,
  XCircle,
  Clock,
  Calendar,
  Film,
  Tv,
  ArrowUpDown,
} from "lucide-react"
import type { Request } from "@/types"

const statusColors = {
  pending: "bg-yellow-500/20 text-yellow-700 dark:text-yellow-300",
  approved: "bg-blue-500/20 text-blue-700 dark:text-blue-300", 
  processing: "bg-purple-500/20 text-purple-700 dark:text-purple-300",
  denied: "bg-red-500/20 text-red-700 dark:text-red-300",
  failed: "bg-red-600/20 text-red-800 dark:text-red-400",
  fulfilled: "bg-green-500/20 text-green-700 dark:text-green-300",
}

const statusIcons = {
  pending: Clock,
  approved: CheckCircle,
  processing: Clock,
  denied: XCircle,
  failed: XCircle,
  fulfilled: CheckCircle,
}

export const createUserRequestsColumns = (navigate: (path: string) => void): ColumnDef<Request>[] => [
  {
    accessorKey: "title",
    header: ({ column }) => {
      return (
        <Button
          variant="ghost"
          onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
          className="h-8 px-2"
        >
          Content
          <ArrowUpDown className="ml-2 h-4 w-4" />
        </Button>
      )
    },
    cell: ({ row }) => {
      const request = row.original
      
      const handlePosterClick = () => {
        if (request.tmdb_id && request.media_type) {
          navigate(`/requests/${request.media_type}/${request.tmdb_id}/details`)
        }
      }

      return (
        <div className="flex items-center gap-3">
          {request.poster_url && (
            <img
              src={request.poster_url}
              alt={request.title}
              className="w-12 h-18 object-cover rounded cursor-pointer hover:opacity-80 transition-opacity"
              onClick={handlePosterClick}
            />
          )}
          <div>
            <div className="font-medium text-foreground">{request.title}</div>
            {request.notes && (
              <div className="text-sm text-muted-foreground mt-1 truncate max-w-[300px]">
                {request.notes}
              </div>
            )}
          </div>
        </div>
      )
    },
  },
  {
    accessorKey: "media_type",
    header: "Type",
    cell: ({ row }) => {
      const mediaType = row.getValue("media_type") as string
      return (
        <div className="flex items-center gap-2">
          {mediaType === 'movie' ? (
            <Film className="w-4 h-4 text-muted-foreground" />
          ) : (
            <Tv className="w-4 h-4 text-muted-foreground" />
          )}
          <span className="capitalize">{mediaType === "tv" ? "Series" : mediaType}</span>
        </div>
      )
    },
  },
  {
    accessorKey: "status",
    header: ({ column }) => {
      return (
        <Button
          variant="ghost"
          onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
          className="h-8 px-2"
        >
          Status
          <ArrowUpDown className="ml-2 h-4 w-4" />
        </Button>
      )
    },
    cell: ({ row }) => {
      const status = row.getValue("status") as keyof typeof statusColors
      const StatusIcon = statusIcons[status] || Clock
      
      return (
        <Badge
          variant="secondary"
          className={statusColors[status]}
        >
          <StatusIcon className="w-3 h-3 mr-1" />
          {status ? status.charAt(0).toUpperCase() + status.slice(1) : 'Unknown'}
        </Badge>
      )
    },
    filterFn: (row, id, value) => {
      return value.includes(row.getValue(id))
    },
  },
  {
    accessorKey: "created_at",
    header: ({ column }) => {
      return (
        <Button
          variant="ghost"
          onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
          className="h-8 px-2"
        >
          Requested
          <ArrowUpDown className="ml-2 h-4 w-4" />
        </Button>
      )
    },
    cell: ({ row }) => {
      const date = row.getValue("created_at") as string
      return (
        <div className="flex items-center gap-2 text-muted-foreground">
          <Calendar className="w-4 h-4" />
          <span>{new Date(date).toLocaleDateString()}</span>
        </div>
      )
    },
  },
  {
    accessorKey: "fulfilled_at",
    header: "Fulfilled",
    cell: ({ row }) => {
      const date = row.getValue("fulfilled_at") as string | null
      if (!date) return <span className="text-muted-foreground">-</span>
      
      return (
        <div className="flex items-center gap-2 text-muted-foreground">
          <Calendar className="w-4 h-4" />
          <span>{new Date(date).toLocaleDateString()}</span>
        </div>
      )
    },
  },
]