import { api } from './api';
import type {
  Invitation,
  CreateInvitationRequest,
  AcceptInvitationRequest,
  InvitationStats
} from '../types';

// Admin invitation management
export const invitationsApi = {
  // Create a new invitation
  createInvitation: async (request: CreateInvitationRequest): Promise<{ invitation: Invitation; invite_url: string }> => {
    const response = await api.post<{ invitation: Invitation; invite_url: string }>('/invitations', request);
    return response.data;
  },

  // Get all invitations (admin only)
  getAllInvitations: async (): Promise<Invitation[]> => {
    const response = await api.get<Invitation[]>('/invitations');
    return response.data;
  },

  // Get invitation statistics (admin only)
  getInvitationStats: async (): Promise<InvitationStats> => {
    const response = await api.get<InvitationStats>('/invitations/stats');
    return response.data;
  },

  // Cancel an invitation (admin only)
  cancelInvitation: async (invitationId: number): Promise<{ message: string; invitation: Partial<Invitation> }> => {
    const response = await api.put<{ message: string; invitation: Partial<Invitation> }>(`/invitations/${invitationId}/cancel`, {});
    return response.data;
  },

  // Delete an invitation (admin only)
  deleteInvitation: async (invitationId: number): Promise<{ message: string }> => {
    await api.delete(`/invitations/${invitationId}`);
    return { message: 'Invitation deleted successfully' };
  },

  // Get invitation by token (public)
  getInvitationByToken: async (token: string): Promise<Invitation> => {
    const response = await api.get<Invitation>(`/invitations/accept/${token}`);
    return response.data;
  },

  // Accept an invitation (public)
  acceptInvitation: async (request: AcceptInvitationRequest): Promise<{ message: string; user: any }> => {
    const response = await api.post<{ message: string; user: any }>('/invitations/accept', request);
    return response.data;
  },

  // Get invitation link (admin only)
  getInvitationLink: async (invitationId: number): Promise<{ invite_url: string }> => {
    const response = await api.get<{ invite_url: string }>(`/invitations/${invitationId}/link`);
    return response.data;
  }
};