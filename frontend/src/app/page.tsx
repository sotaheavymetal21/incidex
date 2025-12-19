import { redirect } from "next/navigation";

export default function Home() {
  // ルートアクセス時はインシデント一覧へリダイレクト
  redirect("/incidents");
}
