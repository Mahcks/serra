import React, { useState, useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { analyticsApi } from '../../lib/analytics-api';
import { mountedDrivesApi } from '../../lib/api';
import { getErrorMessage } from '../../utils/errorHandling';
import type { AnalyticsOverview, DriveAlert } from '../../types/analytics';
import type { CreateMountedDriveRequest } from '../../types';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '../../components/ui/tabs';
import { OverviewTab } from '../../components/admin/analytics/OverviewTab';
import { StorageTab } from '../../components/admin/analytics/StorageTab';
import { RequestsTab } from '../../components/admin/analytics/RequestsTab';
import { WatchTab } from '../../components/admin/analytics/WatchTab';
import { BarChart3, HardDrive, Film, Users } from 'lucide-react';

const AnalyticsPage: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const [overview, setOverview] = useState<AnalyticsOverview | null>(null);
  const [watchAnalytics, setWatchAnalytics] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showAddDriveDialog, setShowAddDriveDialog] = useState(false);
  const [addDriveForm, setAddDriveForm] = useState<CreateMountedDriveRequest>({
    name: '',
    mount_path: ''
  });
  const [addDriveLoading, setAddDriveLoading] = useState(false);
  const [systemDrives, setSystemDrives] = useState<any[]>([]);
  const [showThresholdDialog, setShowThresholdDialog] = useState(false);
  const [thresholdForm, setThresholdForm] = useState<{
    driveId: string;
    driveName: string;
    monitoring_enabled: boolean;
    warning_threshold: number;
    critical_threshold: number;
    growth_rate_threshold: number;
  }>({
    driveId: '',
    driveName: '',
    monitoring_enabled: true,
    warning_threshold: 80,
    critical_threshold: 95,
    growth_rate_threshold: 50
  });
  const [thresholdLoading, setThresholdLoading] = useState(false);

  // Determine current tab from URL
  const getCurrentTab = () => {
    const path = location.pathname;
    if (path.endsWith('/storage')) return 'storage';
    if (path.endsWith('/requests')) return 'requests';
    if (path.endsWith('/watch')) return 'watch';
    return 'overview';
  };

  const currentTab = getCurrentTab();

  // Handle tab changes and URL navigation
  const handleTabChange = (tab: string) => {
    const basePath = '/admin/analytics';
    const newPath = tab === 'overview' ? basePath : `${basePath}/${tab}`;
    navigate(newPath);
  };

  useEffect(() => {
    loadAnalytics();
  }, []);

  const loadAnalytics = async () => {
    try {
      setLoading(true);
      const [overviewData, watchData] = await Promise.all([
        analyticsApi.getOverview(),
        analyticsApi.getWatchAnalytics(15)
      ]);
      
      // Mock storage pools data for UI testing (toggle with VITE_MOCK_STORAGE_POOLS=true)
      if (import.meta.env.VITE_MOCK_STORAGE_POOLS === 'true' && (!overviewData.storage_pools || overviewData.storage_pools.length === 0)) {
        overviewData.storage_pools = [
          {
            name: "tank",
            type: "zfs",
            health: "ONLINE",
            status: "ONLINE",
            total_size: 10995116277760, // 10TB
            used_size: 5497558138880,  // 5TB
            available_size: 5497558138880, // 5TB
            usage_percentage: 50.0,
            redundancy: "raidz2",
            last_checked: new Date().toISOString(),
            devices: [
              { name: "sda1", path: "/dev/sda1", status: "ONLINE", health: "healthy", read_errors: 0, write_errors: 0, checksum_errors: 0 },
              { name: "sdb1", path: "/dev/sdb1", status: "ONLINE", health: "healthy", read_errors: 0, write_errors: 1, checksum_errors: 0 },
              { name: "sdc1", path: "/dev/sdc1", status: "ONLINE", health: "healthy", read_errors: 0, write_errors: 0, checksum_errors: 0 },
              { name: "sdd1", path: "/dev/sdd1", status: "ONLINE", health: "healthy", read_errors: 0, write_errors: 0, checksum_errors: 0 }
            ]
          },
          {
            name: "backup",
            type: "zfs", 
            health: "DEGRADED",
            status: "DEGRADED",
            total_size: 2199023255552, // 2TB
            used_size: 879609302220,  // 800GB
            available_size: 1319413953332, // 1.2TB
            usage_percentage: 40.0,
            redundancy: "mirror",
            last_checked: new Date().toISOString(),
            devices: [
              { name: "sde1", path: "/dev/sde1", status: "ONLINE", health: "healthy", read_errors: 0, write_errors: 0, checksum_errors: 0 },
              { name: "sdf1", path: "/dev/sdf1", status: "DEGRADED", health: "degraded", read_errors: 2, write_errors: 0, checksum_errors: 1 }
            ]
          },
          {
            name: "array1",
            type: "unraid",
            health: "healthy", 
            status: "active",
            total_size: 21990232555520, // 20TB
            used_size: 17592186044416,  // 16TB
            available_size: 4398046511104, // 4TB
            usage_percentage: 80.0,
            redundancy: "Parity: 2",
            last_checked: new Date().toISOString(),
            devices: [
              { name: "sdb", path: "/dev/sdb", status: "active", health: "healthy", read_errors: 0, write_errors: 0, checksum_errors: 0 },
              { name: "sdc", path: "/dev/sdc", status: "active", health: "healthy", read_errors: 0, write_errors: 0, checksum_errors: 0 },
              { name: "sdd", path: "/dev/sdd", status: "active", health: "healthy", read_errors: 0, write_errors: 0, checksum_errors: 0 }
            ]
          }
        ];
      }
      
      setOverview(overviewData);
      setWatchAnalytics(watchData);
      setError(null);
    } catch (err) {
      const errorMessage = getErrorMessage(err);
      setError(errorMessage);
      console.error('Failed to load analytics:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleAcknowledgeAlert = async (alertId: number) => {
    try {
      await analyticsApi.acknowledgeDriveAlert(alertId);
      await loadAnalytics();
    } catch (err) {
      const errorMessage = getErrorMessage(err);
      console.error('Failed to acknowledge alert:', errorMessage, err);
    }
  };

  const handleShowThresholdDialog = (drive: any) => {
    setThresholdForm({
      driveId: drive.drive_id,
      driveName: drive.name || `Drive ${drive.drive_id}`,
      monitoring_enabled: drive.monitoring_enabled ?? true,
      warning_threshold: drive.warning_threshold ?? 80,
      critical_threshold: drive.critical_threshold ?? 95,
      growth_rate_threshold: drive.growth_rate_threshold ?? 50
    });
    setShowThresholdDialog(true);
  };

  const handleUpdateThresholds = async () => {
    if (!thresholdForm.driveId) return;

    try {
      setThresholdLoading(true);
      await mountedDrivesApi.updateDriveThresholds(thresholdForm.driveId, {
        monitoring_enabled: thresholdForm.monitoring_enabled,
        warning_threshold: thresholdForm.warning_threshold,
        critical_threshold: thresholdForm.critical_threshold,
        growth_rate_threshold: thresholdForm.growth_rate_threshold
      });
      
      setShowThresholdDialog(false);
      loadAnalytics();
    } catch (err) {
      const errorMessage = getErrorMessage(err);
      console.error('Failed to update thresholds:', err);
      setError(errorMessage);
    } finally {
      setThresholdLoading(false);
    }
  };

  const handleDeleteDrive = async (driveId: string, driveName: string) => {
    if (!confirm(`Are you sure you want to delete "${driveName}"? This action cannot be undone.`)) {
      return;
    }

    try {
      await mountedDrivesApi.deleteMountedDrive(driveId);
      loadAnalytics();
    } catch (err) {
      const errorMessage = getErrorMessage(err);
      console.error('Failed to delete drive:', err);
      setError(errorMessage);
    }
  };

  const loadSystemDrives = async () => {
    try {
      const drives = await mountedDrivesApi.getSystemDrives();
      setSystemDrives(drives);
    } catch (err) {
      const errorMessage = getErrorMessage(err);
      console.error('Failed to load system drives:', errorMessage, err);
    }
  };

  const handleAddDrive = async () => {
    if (!addDriveForm.name.trim() || !addDriveForm.mount_path.trim()) {
      return;
    }

    try {
      setAddDriveLoading(true);
      await mountedDrivesApi.createMountedDrive(addDriveForm);
      
      setAddDriveForm({ name: '', mount_path: '' });
      setShowAddDriveDialog(false);
      
      await loadAnalytics();
    } catch (err) {
      const errorMessage = getErrorMessage(err);
      console.error('Failed to add drive:', errorMessage, err);
    } finally {
      setAddDriveLoading(false);
    }
  };

  const handleShowAddDriveDialog = () => {
    setShowAddDriveDialog(true);
    loadSystemDrives();
  };

  const formatBytes = (bytes: number | undefined | null) => {
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    if (!bytes || bytes === 0 || isNaN(bytes)) return '0 Bytes';
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i];
  };

  const formatPercentage = (percentage: number | undefined | null) => {
    if (percentage === undefined || percentage === null || isNaN(percentage)) {
      return '0.0%';
    }
    return `${percentage.toFixed(1)}%`;
  };

  const getAlertTypeColor = (alertType: string) => {
    switch (alertType) {
      case 'usage_threshold':
        return 'bg-red-100 text-red-800 border-red-200';
      case 'growth_rate':
        return 'bg-yellow-100 text-yellow-800 border-yellow-200';
      case 'projected_full':
        return 'bg-orange-100 text-orange-800 border-orange-200';
      default:
        return 'bg-gray-100 text-gray-800 border-gray-200';
    }
  };

  const getAlertPriority = (alert: DriveAlert) => {
    if (alert.alert_type === 'usage_threshold' && alert.current_value >= 95) {
      return 'CRITICAL';
    } else if (alert.alert_type === 'usage_threshold' && alert.current_value >= 80) {
      return 'WARNING';
    }
    return 'INFO';
  };

  if (loading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="animate-pulse">
          <div className="h-8 bg-muted rounded w-1/4 mb-6"></div>
          <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
            {[...Array(4)].map((_, i) => (
              <div key={i} className="bg-card p-6 rounded-lg shadow border">
                <div className="h-4 bg-muted rounded w-3/4 mb-2"></div>
                <div className="h-8 bg-muted rounded w-1/2"></div>
              </div>
            ))}
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="bg-destructive/15 border border-destructive/20 text-destructive px-4 py-3 rounded">
          {error}
          <button
            onClick={loadAnalytics}
            className="ml-4 bg-destructive text-destructive-foreground px-3 py-1 rounded text-sm hover:bg-destructive/90"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  if (!overview) return null;

  return (
    <div className="container mx-auto">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-foreground mb-2">System Analytics</h1>
        <p className="text-muted-foreground">Drive monitoring, request analytics, and system health</p>
      </div>

      <Tabs value={currentTab} onValueChange={handleTabChange} className="w-full">
        <TabsList className="grid w-full grid-cols-4 mb-8">
          <TabsTrigger value="overview" className="flex items-center gap-2">
            <BarChart3 className="h-4 w-4" />
            Overview
          </TabsTrigger>
          <TabsTrigger value="storage" className="flex items-center gap-2">
            <HardDrive className="h-4 w-4" />
            Storage
          </TabsTrigger>
          <TabsTrigger value="requests" className="flex items-center gap-2">
            <Film className="h-4 w-4" />
            Requests
          </TabsTrigger>
          <TabsTrigger value="watch" className="flex items-center gap-2">
            <Users className="h-4 w-4" />
            Watch
          </TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="mt-0">
          <OverviewTab
            overview={overview}
            watchAnalytics={watchAnalytics}
            formatBytes={formatBytes}
            formatPercentage={formatPercentage}
            handleAcknowledgeAlert={handleAcknowledgeAlert}
            getAlertTypeColor={getAlertTypeColor}
            getAlertPriority={getAlertPriority}
          />
        </TabsContent>

        <TabsContent value="storage" className="mt-0">
          <StorageTab
            overview={overview}
            formatBytes={formatBytes}
            formatPercentage={formatPercentage}
            handleShowThresholdDialog={handleShowThresholdDialog}
            handleDeleteDrive={handleDeleteDrive}
            handleShowAddDriveDialog={handleShowAddDriveDialog}
          />
        </TabsContent>

        <TabsContent value="requests" className="mt-0">
          <RequestsTab />
        </TabsContent>

        <TabsContent value="watch" className="mt-0">
          <WatchTab />
        </TabsContent>
      </Tabs>

      {/* Add Drive Dialog */}
      {showAddDriveDialog && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-card rounded-lg p-6 w-full max-w-md mx-4">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold text-foreground">Add Drive</h3>
              <button
                onClick={() => setShowAddDriveDialog(false)}
                className="text-muted-foreground hover:text-foreground"
              >
                <span className="sr-only">Close</span>
                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>

            <div className="space-y-4">
              <div>
                <label htmlFor="drive-name" className="block text-sm font-medium text-foreground mb-1">
                  Drive Name
                </label>
                <input
                  id="drive-name"
                  type="text"
                  value={addDriveForm.name}
                  onChange={(e) => setAddDriveForm(prev => ({ ...prev, name: e.target.value }))}
                  className="w-full px-3 py-2 border border-input rounded-md bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent"
                  placeholder="Enter drive name"
                />
              </div>

              <div>
                <label htmlFor="mount-path" className="block text-sm font-medium text-foreground mb-1">
                  Mount Path
                </label>
                <input
                  id="mount-path"
                  type="text"
                  value={addDriveForm.mount_path}
                  onChange={(e) => setAddDriveForm(prev => ({ ...prev, mount_path: e.target.value }))}
                  className="w-full px-3 py-2 border border-input rounded-md bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent"
                  placeholder="/path/to/mount"
                />
              </div>

              {systemDrives.length > 0 && (
                <div>
                  <label className="block text-sm font-medium text-foreground mb-2">
                    Available System Drives
                  </label>
                  <div className="space-y-2 max-h-32 overflow-y-auto">
                    {systemDrives.map((drive, index) => (
                      <button
                        key={index}
                        onClick={() => setAddDriveForm(prev => ({ 
                          ...prev, 
                          mount_path: drive.mount_path,
                          name: prev.name || drive.name 
                        }))}
                        className="w-full text-left p-2 bg-muted hover:bg-muted/80 rounded border text-sm"
                      >
                        <div className="font-medium">{drive.name}</div>
                        <div className="text-muted-foreground">{drive.mount_path}</div>
                      </button>
                    ))}
                  </div>
                </div>
              )}
            </div>

            <div className="flex items-center justify-end gap-3 mt-6">
              <button
                onClick={() => setShowAddDriveDialog(false)}
                className="px-4 py-2 text-foreground bg-secondary rounded-md hover:bg-secondary/80 transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleAddDrive}
                disabled={addDriveLoading || !addDriveForm.name.trim() || !addDriveForm.mount_path.trim()}
                className="px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              >
                {addDriveLoading ? 'Adding...' : 'Add Drive'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Threshold Configuration Dialog */}
      {showThresholdDialog && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-card rounded-lg p-6 w-full max-w-md mx-4">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold text-foreground">
                Configure Thresholds: {thresholdForm.driveName}
              </h3>
              <button
                onClick={() => setShowThresholdDialog(false)}
                className="text-muted-foreground hover:text-foreground"
              >
                <span className="sr-only">Close</span>
                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>

            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <label htmlFor="monitoring-enabled" className="text-sm font-medium text-foreground">
                  Enable Monitoring
                </label>
                <input
                  id="monitoring-enabled"
                  type="checkbox"
                  checked={thresholdForm.monitoring_enabled}
                  onChange={(e) => setThresholdForm(prev => ({ ...prev, monitoring_enabled: e.target.checked }))}
                  className="rounded border-input"
                />
              </div>

              <div>
                <label htmlFor="warning-threshold" className="block text-sm font-medium text-foreground mb-1">
                  Warning Threshold (%)
                </label>
                <input
                  id="warning-threshold"
                  type="number"
                  min="0"
                  max="100"
                  value={thresholdForm.warning_threshold}
                  onChange={(e) => setThresholdForm(prev => ({ ...prev, warning_threshold: Number(e.target.value) }))}
                  className="w-full px-3 py-2 border border-input rounded-md bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent"
                />
              </div>

              <div>
                <label htmlFor="critical-threshold" className="block text-sm font-medium text-foreground mb-1">
                  Critical Threshold (%)
                </label>
                <input
                  id="critical-threshold"
                  type="number"
                  min="0"
                  max="100"
                  value={thresholdForm.critical_threshold}
                  onChange={(e) => setThresholdForm(prev => ({ ...prev, critical_threshold: Number(e.target.value) }))}
                  className="w-full px-3 py-2 border border-input rounded-md bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent"
                />
              </div>

              <div>
                <label htmlFor="growth-rate-threshold" className="block text-sm font-medium text-foreground mb-1">
                  Growth Rate Threshold (GB/day)
                </label>
                <input
                  id="growth-rate-threshold"
                  type="number"
                  min="0"
                  step="0.1"
                  value={thresholdForm.growth_rate_threshold}
                  onChange={(e) => setThresholdForm(prev => ({ ...prev, growth_rate_threshold: Number(e.target.value) }))}
                  className="w-full px-3 py-2 border border-input rounded-md bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent"
                />
              </div>
            </div>

            <div className="flex items-center justify-end gap-3 mt-6">
              <button
                onClick={() => setShowThresholdDialog(false)}
                className="px-4 py-2 text-foreground bg-secondary rounded-md hover:bg-secondary/80 transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleUpdateThresholds}
                disabled={thresholdLoading}
                className="px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              >
                {thresholdLoading ? 'Updating...' : 'Update Thresholds'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default AnalyticsPage;