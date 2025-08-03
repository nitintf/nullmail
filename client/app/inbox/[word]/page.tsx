import { InboxContainer } from "@/components/inbox-container";

export default async function InboxPage({ params }: { params: Promise<{ word: string }> }) {
  const { word } = await params;
  const emailAddress = `${word}@nullmail.nitin.sh`;

  return <InboxContainer emailAddress={emailAddress} />;
} 