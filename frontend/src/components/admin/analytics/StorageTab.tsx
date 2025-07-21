import React from 'react';
import type { AnalyticsOverview } from '../../../types/analytics';

interface StorageTabProps {
  overview: AnalyticsOverview;
  formatBytes: (bytes: number | undefined | null) => string;
  formatPercentage: (percentage: number | undefined | null) => string;
  handleShowThresholdDialog: (drive: any) => void;
  handleDeleteDrive: (driveId: string, driveName: string) => void;
  handleShowAddDriveDialog: () => void;
}

export const StorageTab: React.FC<StorageTabProps> = ({
  overview,
  formatBytes,
  formatPercentage,
  handleShowThresholdDialog,
  handleDeleteDrive,
  handleShowAddDriveDialog
}) => {
  return (
    <>
      {/* Storage Pools */}
      {overview.storage_pools !== undefined && (
        <div className="bg-card rounded-lg shadow border mb-8">
          <div className="px-6 py-4 border-b border-border">
            <h2 className="text-xl font-semibold text-foreground">Storage Pools</h2>
          </div>
          <div className="p-6">
            {overview.storage_pools && overview.storage_pools.length > 0 ? (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {overview.storage_pools.map((pool, index) => (
                <div key={index} className="border border-border rounded-lg p-4">
                  <div className="flex items-center justify-between mb-2">
                    <div className="flex items-center gap-2">
                      <h3 className="font-medium text-foreground">{pool.name}</h3>
                      <span className={`text-xs px-2 py-1 rounded ${
                        pool.type === 'zfs' ? 'bg-blue-100 text-blue-800' :
                        pool.type === 'unraid' ? 'bg-purple-100 text-purple-800' :
                        'bg-gray-100 text-gray-800'
                      }`}>
                        {pool.type.toUpperCase()}
                      </span>
                    </div>
                    <span className={`text-sm px-2 py-1 rounded ${
                      pool.health === 'ONLINE' || pool.health === 'healthy' ? 'bg-green-100 text-green-800' :
                      pool.health === 'DEGRADED' ? 'bg-yellow-100 text-yellow-800' :
                      'bg-red-100 text-red-800'
                    }`}>
                      {pool.health}
                    </span>
                  </div>
                  
                  <div className="space-y-2 text-sm text-muted-foreground">
                    <div className="flex justify-between">
                      <span>Status:</span>
                      <span className="font-mono text-xs">{pool.status}</span>
                    </div>
                    <div className="flex justify-between">
                      <span>Total:</span>
                      <span>{formatBytes(pool.total_size)}</span>
                    </div>
                    <div className="flex justify-between">
                      <span>Used:</span>
                      <span>{formatBytes(pool.used_size)}</span>
                    </div>
                    <div className="flex justify-between">
                      <span>Available:</span>
                      <span>{formatBytes(pool.available_size)}</span>
                    </div>
                    {pool.redundancy && (
                      <div className="flex justify-between">
                        <span>Redundancy:</span>
                        <span>{pool.redundancy}</span>
                      </div>
                    )}
                    <div className="flex justify-between text-xs">
                      <span>Updated:</span>
                      <span>{new Date(pool.last_checked).toLocaleString()}</span>
                    </div>
                  </div>
                  
                  {/* Usage bar */}
                  <div className="mt-3">
                    <div className="bg-muted rounded-full h-2">
                      <div
                        className={`h-2 rounded-full ${
                          pool.usage_percentage >= 95 ? 'bg-red-500' :
                          pool.usage_percentage >= 80 ? 'bg-yellow-500' :
                          'bg-green-500'
                        }`}
                        style={{ width: `${Math.min(pool.usage_percentage, 100)}%` }}
                      ></div>
                    </div>
                    <div className="text-center text-xs text-muted-foreground mt-1">
                      {formatPercentage(pool.usage_percentage)}
                    </div>
                  </div>
                  
                  {/* Device count */}
                  {pool.devices && pool.devices.length > 0 && (
                    <div className="mt-3 pt-3 border-t border-border">
                      <div className="text-xs text-muted-foreground">
                        {pool.devices.length} device{pool.devices.length !== 1 ? 's' : ''}
                        {pool.devices.some((d: any) => (d.read_errors || 0) > 0 || (d.write_errors || 0) > 0 || (d.checksum_errors || 0) > 0) && (
                          <span className="ml-2 text-yellow-600">⚠ Errors detected</span>
                        )}
                      </div>
                    </div>
                  )}
                </div>
                ))}
              </div>
            ) : (
              <div className="text-center py-8 text-muted-foreground">
                <div className="mb-4">
                  <svg className="w-16 h-16 mx-auto text-muted-foreground/50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
                  </svg>
                </div>
                <p className="text-lg font-medium mb-2">No Storage Pools Detected</p>
                <p className="text-sm">ZFS pools and UnRAID arrays will appear here when detected.</p>
                <p className="text-xs mt-2 text-muted-foreground/75">
                  Supported: ZFS pools, UnRAID arrays
                </p>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Drive Usage Overview */}
      <div className="bg-card rounded-lg shadow border mb-8">
        <div className="px-6 py-4 border-b border-border flex items-center justify-between">
          <h2 className="text-xl font-semibold text-foreground">Drive Usage</h2>
          <button
            onClick={handleShowAddDriveDialog}
            className="bg-primary text-primary-foreground px-4 py-2 rounded-lg text-sm hover:bg-primary/90 transition-colors"
          >
            Add Drive
          </button>
        </div>
        <div className="p-6">
          {overview.drive_usage && overview.drive_usage.length > 0 ? (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {overview.drive_usage.map((drive: any, index) => (
                <div key={index} className={`border border-border rounded-lg p-4 flex flex-col ${!drive.has_data ? 'opacity-75 bg-muted/50' : ''}`}>
                <div className="flex items-center justify-between mb-2">
                  <div className="flex items-center gap-2">
                    <h3 className="font-medium text-foreground">
                      {drive.name || `Drive ${index + 1}`}
                    </h3>
                    {!drive.has_data && (
                      <span className="text-xs bg-yellow-100 text-yellow-800 px-2 py-1 rounded">
                        Offline
                      </span>
                    )}
                  </div>
                  {drive.has_data && (
                    <span className={`text-sm px-2 py-1 rounded ${
                      (drive.usage_percentage ?? 0) >= 95 ? 'bg-red-100 text-red-800' :
                      (drive.usage_percentage ?? 0) >= 80 ? 'bg-yellow-100 text-yellow-800' :
                      'bg-green-100 text-green-800'
                    }`}>
                      {formatPercentage(drive.usage_percentage)}
                    </span>
                  )}
                </div>
                
                {/* Content area - flex-grow to push buttons to bottom */}
                <div className="flex-grow">
                  <div className="space-y-2 text-sm text-muted-foreground">
                    {drive.mount_path && (
                      <div className="flex justify-between">
                        <span>Mount:</span>
                        <span className="font-mono text-xs">{drive.mount_path}</span>
                      </div>
                    )}
                    {drive.has_data ? (
                      <>
                        <div className="flex justify-between">
                          <span>Total:</span>
                          <span>{formatBytes(drive.total_size ?? 0)}</span>
                        </div>
                        <div className="flex justify-between">
                          <span>Used:</span>
                          <span>{formatBytes(drive.used_size ?? 0)}</span>
                        </div>
                        <div className="flex justify-between">
                          <span>Free:</span>
                          <span>{formatBytes(drive.available_size ?? 0)}</span>
                        </div>
                        {(drive.growth_rate_gb_per_day ?? 0) > 0 && (
                          <div className="flex justify-between">
                            <span>Growth:</span>
                            <span>{(drive.growth_rate_gb_per_day ?? 0).toFixed(1)} GB/day</span>
                          </div>
                        )}
                        {drive.recorded_at && (
                          <div className="flex justify-between text-xs">
                            <span>Updated:</span>
                            <span>{new Date(drive.recorded_at).toLocaleString()}</span>
                          </div>
                        )}
                      </>
                    ) : (
                      <div className="text-center py-4 text-muted-foreground">
                        <p>No usage data available</p>
                        <p className="text-xs">Drive not monitored or offline</p>
                      </div>
                    )}
                  </div>
                  
                  {/* Usage bar */}
                  {drive.has_data && (
                    <div className="mt-3">
                      <div className="bg-muted rounded-full h-2">
                        <div
                          className={`h-2 rounded-full ${
                            (drive.usage_percentage ?? 0) >= 95 ? 'bg-red-500' :
                            (drive.usage_percentage ?? 0) >= 80 ? 'bg-yellow-500' :
                            'bg-green-500'
                          }`}
                          style={{ width: `${Math.min(drive.usage_percentage ?? 0, 100)}%` }}
                        ></div>
                      </div>
                    </div>
                  )}
                </div>
                
                {/* Action buttons - always at bottom */}
                <div className="mt-3 pt-3 border-t border-border space-y-2">
                  <button
                    onClick={() => handleShowThresholdDialog(drive)}
                    className="w-full bg-secondary text-secondary-foreground px-3 py-2 rounded text-sm hover:bg-secondary/80 transition-colors"
                  >
                    Configure Thresholds
                  </button>
                  <button
                    onClick={() => handleDeleteDrive(drive.drive_id, drive.name || `Drive ${index + 1}`)}
                    className="w-full bg-destructive text-destructive-foreground px-3 py-2 rounded text-sm hover:bg-destructive/90 transition-colors"
                  >
                    Delete Drive
                  </button>
                </div>
              </div>
            ))}
            </div>
          ) : (
            <div className="text-center py-8 text-muted-foreground">
              <p>No drive usage data available</p>
            </div>
          )}
        </div>
      </div>
    </>
  );
};