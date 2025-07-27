import {
  type ColumnDef,
  flexRender,
  getCoreRowModel,
  useReactTable,
  getSortedRowModel,
  type SortingState,
  getFilteredRowModel,
  type ColumnFiltersState,
} from "@tanstack/react-table";

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useState } from "react";
import { Search, X } from "lucide-react";

interface DataTableProps<TData, TValue> {
  columns: ColumnDef<TData, TValue>[];
  data: TData[];
  compact?: boolean;
}

export function InvitationDataTable<TData, TValue>({
  columns,
  data,
  compact = false,
}: DataTableProps<TData, TValue>) {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const [globalFilter, setGlobalFilter] = useState("");

  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    onSortingChange: setSorting,
    getSortedRowModel: getSortedRowModel(),
    onColumnFiltersChange: setColumnFilters,
    getFilteredRowModel: getFilteredRowModel(),
    onGlobalFilterChange: setGlobalFilter,
    globalFilterFn: "includesString",
    state: {
      sorting,
      columnFilters,
      globalFilter,
    },
  });

  const statusFilter = table.getColumn("status")?.getFilterValue() as string;

  return (
    <div className="space-y-4">
      {/* Filters - hide in compact mode */}
      {!compact && (
        <>
          <div className="flex items-center space-x-2">
            <div className="flex items-center space-x-2 flex-1">
              <Search className="h-4 w-4 text-gray-400" />
              <Input
                placeholder="Search invitations..."
                value={globalFilter ?? ""}
                onChange={(event) => setGlobalFilter(String(event.target.value))}
                className="max-w-sm"
              />
            </div>
            
            <Select
              value={statusFilter ?? "all"}
              onValueChange={(value) =>
                table.getColumn("status")?.setFilterValue(value === "all" ? "" : value)
              }
            >
              <SelectTrigger className="w-[150px]">
                <SelectValue placeholder="Status" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Status</SelectItem>
                <SelectItem value="pending">Pending</SelectItem>
                <SelectItem value="accepted">Accepted</SelectItem>
                <SelectItem value="expired">Expired</SelectItem>
                <SelectItem value="cancelled">Cancelled</SelectItem>
              </SelectContent>
            </Select>

            {(globalFilter || statusFilter) && (
              <Button
                variant="ghost"
                onClick={() => {
                  setGlobalFilter("");
                  table.getColumn("status")?.setFilterValue("");
                }}
                className="h-8 px-2 lg:px-3"
              >
                Reset
                <X className="ml-2 h-4 w-4" />
              </Button>
            )}
          </div>

          {/* Results info */}
          <div className="flex items-center justify-between">
            <div className="text-sm text-gray-600">
              {table.getFilteredRowModel().rows.length} of {table.getCoreRowModel().rows.length} invitation(s)
            </div>
            
            {/* Status counts */}
            <div className="flex items-center space-x-2">
              {Array.isArray(data) && data.length > 0 && ['pending', 'accepted', 'expired', 'cancelled'].map((status) => {
                const count = data.filter((item) => {
                  return typeof item === 'object' && item !== null && 'status' in item && 
                         (item as { status: string }).status === status;
                }).length;
                if (count === 0) return null;
                
                const getStatusColor = (status: string) => {
                  switch (status) {
                    case 'pending':
                      return 'bg-yellow-100 text-yellow-800';
                    case 'accepted':
                      return 'bg-green-100 text-green-800';
                    case 'expired':
                      return 'bg-gray-100 text-gray-800';
                    case 'cancelled':
                      return 'bg-red-100 text-red-800';
                    default:
                      return 'bg-gray-100 text-gray-800';
                  }
                };
                
                return (
                  <Badge key={status} variant="outline" className={getStatusColor(status)}>
                    {count} {status}
                  </Badge>
                );
              })}
            </div>
          </div>
        </>
      )}

      {/* Table */}
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <TableHead key={header.id}>
                    {header.isPlaceholder
                      ? null
                      : flexRender(
                          header.column.columnDef.header,
                          header.getContext()
                        )}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  data-state={row.getIsSelected() && "selected"}
                >
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id}>
                      {flexRender(
                        cell.column.columnDef.cell,
                        cell.getContext()
                      )}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell
                  colSpan={columns.length}
                  className="h-24 text-center"
                >
                  No invitations found.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}