import { redirect } from "next/navigation";

export default function Home() {
  // ルートアクセス時はダッシュボードへリダイレクト
  redirect("/dashboard");
}
