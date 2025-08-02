import { queryOptions } from '@tanstack/react-query';
import { Email } from './types';

export const emailsQueryOptions = (emailAddress: string) => 
  queryOptions({
    queryKey: ['emails', emailAddress],
    queryFn: async (): Promise<Email[]> => {
      const response = await fetch(`/api/emails/${encodeURIComponent(emailAddress)}`);
      
      if (!response.ok) {
        throw new Error(`Failed to fetch emails: ${response.statusText}`);
      }
      
      return response.json();
    },
    refetchInterval: 30000,
    staleTime: 10000,
    gcTime: 5 * 60 * 1000,
  });

export const emailQueryOptions = (emailAddress: string, emailId: string) =>
  queryOptions({
    queryKey: ['email', emailAddress, emailId],
    queryFn: async (): Promise<Email> => {
      const response = await fetch(`/api/emails/${encodeURIComponent(emailAddress)}/${emailId}`);
      
      if (!response.ok) {
        throw new Error(`Failed to fetch email: ${response.statusText}`);
      }
      
      return response.json();
    },
    staleTime: 5 * 60 * 1000,
    gcTime: 10 * 60 * 1000,
  });