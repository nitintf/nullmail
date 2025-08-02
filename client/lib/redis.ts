import { createClient } from 'redis';
import { Email } from './types';

let client: any = null;

export async function getRedisClient() {
  if (!client) {
    client = createClient({
      url: process.env.REDIS_URL || 'redis://localhost:6379',
      password: process.env.REDIS_PASSWORD || 'dev123',
    });

    client.on('error', (err: any) => console.error('Redis Client Error', err));
    
    if (!client.isOpen) {
      await client.connect();
    }
  }
  
  return client;
}

export async function getEmailsForAddress(emailAddress: string): Promise<Email[]> {
  try {
    const redis = await getRedisClient();
    
    // Use the new recipient index approach
    const addressListKey = `emails:${emailAddress}`;
    const emailIds = await redis.lRange(addressListKey, 0, -1);
    
    console.log(`Found ${emailIds.length} emails for ${emailAddress} using recipient index`);
    
    if (!emailIds || emailIds.length === 0) {
      return [];
    }

    // Batch fetch all emails using pipeline for better performance
    const pipeline = redis.multi();
    emailIds.forEach((emailId: string) => {
      pipeline.get(`nullmail:email:${emailId}`);
    });
    
    const results = await pipeline.exec();
    
    if (!results) {
      console.error('Pipeline execution failed');
      return [];
    }

    const emails: Email[] = [];
    results.forEach((result: any, index: number) => {
      if (result) {
        try {
          console.log(result)
          const email = JSON.parse(result as string);
          const emailId = emailIds[index];
          
          emails.push({
            id: emailId,
            from: email.from || 'unknown@example.com',
            subject: email.subject || 'No Subject',
            body: email.body?.text || email.body || 'No content',
            timestamp: email.received_at || email.timestamp || new Date().toISOString(),
            read: email.read || false,
            starred: email.starred || false,
            hasAttachments: email.attachments && email.attachments.length > 0,
            recipients: email.recipients,
            headers: email.headers,
            attachments: email.attachments,
          });
        } catch (parseError) {
          console.error('Error parsing email data:', parseError);
        }
      }
    });

    // Sort by timestamp, newest first
    return emails.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime());
  } catch (error) {
    console.error('Error fetching emails from Redis:', error);
    return [];
  }
}

export async function getEmailById(emailId: string): Promise<Email | null> {
  try {
    const redis = await getRedisClient();
    const emailData = await redis.get(`nullmail:email:${emailId}`);
    
    if (!emailData) {
      return null;
    }

    const email = JSON.parse(emailData);
    return {
      id: emailId,
      from: email.from || 'unknown@example.com',
      subject: email.subject || 'No Subject',
      body: email.body?.text || email.body || 'No content',
      timestamp: email.received_at || email.timestamp || new Date().toISOString(),
      read: email.read || false,
      starred: email.starred || false,
      hasAttachments: email.attachments && email.attachments.length > 0,
      recipients: email.recipients,
      headers: email.headers,
      attachments: email.attachments,
    };
  } catch (error) {
    console.error('Error fetching email by ID from Redis:', error);
    return null;
  }
}