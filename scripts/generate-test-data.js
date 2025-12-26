const API_URL = process.env.API_URL || 'http://localhost:8080/api';

// ランダムな要素を選択
const randomElement = (arr) => arr[Math.floor(Math.random() * arr.length)];

// ランダムな整数
const randomInt = (min, max) => Math.floor(Math.random() * (max - min + 1)) + min;

// ランダムな日付（過去90日以内）
const randomPastDate = (daysAgo = 90) => {
  const date = new Date();
  date.setDate(date.getDate() - randomInt(0, daysAgo));
  date.setHours(randomInt(0, 23), randomInt(0, 59), randomInt(0, 59));
  return date.toISOString();
};

// ログイン
async function login(email, password) {
  const response = await fetch(`${API_URL}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password })
  });

  if (!response.ok) {
    throw new Error(`Login failed: ${response.status}`);
  }

  const data = await response.json();
  return data.token;
}

// ユーザー作成
async function createUsers(token, existingCount, count = 20) {
  console.log(`\n=== Creating ${count} users (starting from ${existingCount + 1}) ===`);
  const departments = ['開発部', 'インフラ部', 'セキュリティ部', 'QA部', 'プロダクト部', 'デザイン部', '営業部', 'カスタマーサポート'];
  const roles = ['admin', 'editor', 'viewer'];
  const createdUsers = [];

  for (let i = 1; i <= count; i++) {
    try {
      const userNum = existingCount + i;
      const uniqueId = Date.now() + i * 1000 + Math.floor(Math.random() * 1000); // よりユニークな値
      const user = {
        name: `テストユーザー${userNum}`,
        email: `testuser${uniqueId}@example.com`, // タイムスタンプベースでメールもユニークに
        password: 'password123',
        employee_number: `TEST-${uniqueId}`, // 完全にユニークな番号
        department: randomElement(departments),
        role: i <= 3 ? 'admin' : i <= 8 ? 'editor' : 'viewer'
      };

      const response = await fetch(`${API_URL}/users`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify(user)
      });

      if (response.ok) {
        const created = await response.json();
        createdUsers.push(created);
        console.log(`✓ Created user: ${user.name} (${user.email})`);
      } else {
        const error = await response.json().catch(() => ({ error: 'Unknown error' }));
        console.log(`✗ Failed to create user: ${user.email} - ${error.error || response.statusText}`);
      }
    } catch (err) {
      console.error(`Error creating user ${i}:`, err.message);
    }
  }

  return createdUsers;
}

// タグ作成
async function createTags(token, existingTags = []) {
  console.log(`\n=== Creating tags ===`);
  const existingNames = existingTags.map(t => t.name);
  const tags = [
    { name: 'データベース', color: '#3B82F6' },
    { name: 'API', color: '#10B981' },
    { name: 'フロントエンド', color: '#F59E0B' },
    { name: 'バックエンド', color: '#8B5CF6' },
    { name: 'インフラ', color: '#EF4444' },
    { name: 'セキュリティ', color: '#DC2626' },
    { name: 'パフォーマンス', color: '#F97316' },
    { name: 'バグ', color: '#EC4899' },
    { name: 'デプロイ', color: '#06B6D4' },
    { name: 'ネットワーク', color: '#6366F1' },
    { name: '認証', color: '#14B8A6' },
    { name: 'ストレージ', color: '#84CC16' },
    { name: 'キャッシュ', color: '#A855F7' },
    { name: 'モニタリング', color: '#F43F5E' },
    { name: 'ログ', color: '#0EA5E9' },
    { name: 'メール', color: '#22C55E' },
    { name: 'UI/UX', color: '#FBBF24' },
    { name: 'テスト', color: '#64748B' },
    { name: '本番環境', color: '#DC2626' },
    { name: '開発環境', color: '#10B981' }
  ];

  const createdTags = [];

  for (const tag of tags) {
    // Skip if tag already exists
    if (existingNames.includes(tag.name)) {
      console.log(`- Skipping existing tag: ${tag.name}`);
      continue;
    }

    try {
      const response = await fetch(`${API_URL}/tags`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify(tag)
      });

      if (response.ok) {
        const created = await response.json();
        createdTags.push(created);
        console.log(`✓ Created tag: ${tag.name}`);
      }
    } catch (err) {
      console.error(`Error creating tag ${tag.name}:`, err.message);
    }
  }

  return createdTags;
}

// テンプレート作成
async function createTemplates(token, tags, count = 10) {
  console.log(`\n=== Creating ${count} templates ===`);
  const templates = [
    {
      name: 'データベース障害テンプレート',
      description: 'データベース関連の障害発生時に使用するテンプレート',
      title: 'データベース接続障害',
      content: 'データベースへの接続が失敗しています。以下の対応を実施してください：\n1. データベースサーバーの稼働状況確認\n2. 接続プールの状態確認\n3. ネットワーク疎通確認',
      severity: 'high',
      impact_scope: 'システム全体',
      is_public: true,
      tag_ids: tags.filter(t => ['データベース', 'インフラ'].includes(t.name)).map(t => t.id)
    },
    {
      name: 'API障害テンプレート',
      description: 'API エンドポイントの障害発生時に使用',
      severity: 'high',
      tag_ids: tags.filter(t => ['API', 'バックエンド'].includes(t.name)).map(t => t.id)
    },
    {
      name: 'セキュリティインシデントテンプレート',
      description: 'セキュリティ関連のインシデント用',
      severity: 'critical',
      tag_ids: tags.filter(t => ['セキュリティ'].includes(t.name)).map(t => t.id)
    },
    {
      name: 'パフォーマンス劣化テンプレート',
      description: 'システムパフォーマンスの劣化を検知した際に使用',
      severity: 'medium',
      tag_ids: tags.filter(t => ['パフォーマンス'].includes(t.name)).map(t => t.id)
    },
    {
      name: 'デプロイ失敗テンプレート',
      description: 'デプロイ作業の失敗時に使用',
      severity: 'high',
      tag_ids: tags.filter(t => ['デプロイ', 'インフラ'].includes(t.name)).map(t => t.id)
    },
    {
      name: 'ネットワーク障害テンプレート',
      description: 'ネットワーク関連の障害発生時',
      severity: 'critical',
      tag_ids: tags.filter(t => ['ネットワーク', 'インフラ'].includes(t.name)).map(t => t.id)
    },
    {
      name: '認証エラーテンプレート',
      description: 'ログイン・認証関連のエラー',
      severity: 'high',
      tag_ids: tags.filter(t => ['認証', 'セキュリティ'].includes(t.name)).map(t => t.id)
    },
    {
      name: 'UI/UXバグテンプレート',
      description: 'フロントエンドのバグ報告用',
      severity: 'low',
      tag_ids: tags.filter(t => ['フロントエンド', 'UI/UX', 'バグ'].includes(t.name)).map(t => t.id)
    },
    {
      name: 'ストレージ容量不足テンプレート',
      description: 'ストレージ容量不足の警告',
      severity: 'medium',
      tag_ids: tags.filter(t => ['ストレージ', 'インフラ'].includes(t.name)).map(t => t.id)
    },
    {
      name: '監視アラートテンプレート',
      description: 'モニタリングツールからのアラート',
      severity: 'medium',
      tag_ids: tags.filter(t => ['モニタリング'].includes(t.name)).map(t => t.id)
    }
  ];

  const createdTemplates = [];

  for (const template of templates.slice(0, count)) {
    try {
      // テンプレートに必須フィールドを追加
      const fullTemplate = {
        name: template.name,
        description: template.description || '',
        title: template.title || template.name,
        content: template.content || `${template.description}\n\n対応手順：\n1. 状況確認\n2. 原因調査\n3. 対応実施`,
        severity: template.severity,
        impact_scope: template.impact_scope || '影響範囲を記載',
        is_public: template.is_public !== undefined ? template.is_public : true,
        tag_ids: template.tag_ids || []
      };

      const response = await fetch(`${API_URL}/templates`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify(fullTemplate)
      });

      if (response.ok) {
        const created = await response.json();
        createdTemplates.push(created);
        console.log(`✓ Created template: ${template.name}`);
      } else {
        const error = await response.json().catch(() => ({ error: 'Unknown error' }));
        console.log(`✗ Failed to create template: ${template.name} - ${error.error || response.statusText}`);
      }
    } catch (err) {
      console.error(`Error creating template ${template.name}:`, err.message);
    }
  }

  return createdTemplates;
}

// インシデント作成
async function createIncidents(token, users, tags, count = 50) {
  console.log(`\n=== Creating ${count} incidents ===`);
  const severities = ['critical', 'high', 'medium', 'low'];
  const statuses = ['open', 'investigating', 'resolved', 'closed'];

  const titlePrefixes = [
    'データベース接続エラー',
    'API応答タイムアウト',
    'ログイン失敗率の上昇',
    'メモリ使用率の異常増加',
    'ディスク容量不足',
    'CPU使用率100%継続',
    'ネットワーク遅延発生',
    'キャッシュサーバー停止',
    'デプロイ失敗',
    '不正アクセス検知',
    'バッチ処理の異常終了',
    'メール送信失敗',
    'ファイルアップロードエラー',
    'セッションタイムアウト頻発',
    'レスポンスタイム劣化',
    '404エラー増加',
    '500エラー多発',
    'ページ表示崩れ',
    '決済処理エラー',
    'データ同期失敗'
  ];

  const descriptions = [
    'ユーザーから複数の問い合わせあり。調査を開始。',
    '監視ツールでアラートを検知。影響範囲を確認中。',
    '本番環境で突然発生。再現手順を調査中。',
    '定期メンテナンス後に発生を確認。ロールバックを検討。',
    '特定の条件下でのみ発生する模様。原因を特定中。',
    '外部APIの障害が原因と思われる。ベンダーに問い合わせ中。',
    'サーバーログに異常なエラーログを確認。詳細調査中。',
    '負荷テスト中に発見。本番環境への影響は現時点でなし。',
    '一時的なネットワーク障害が原因。復旧済み。',
    'コードレビューで潜在的な問題を発見。修正パッチを準備中。'
  ];

  const impactScopes = [
    '全ユーザー',
    '一部ユーザー（約10%）',
    '管理者のみ',
    '特定機能のみ',
    '東京リージョンのみ',
    '新規登録ユーザーのみ',
    'プレミアムプラン利用者のみ',
    'モバイルアプリユーザーのみ',
    'API経由のアクセスのみ',
    '影響なし（潜在的バグ）'
  ];

  const createdIncidents = [];

  for (let i = 1; i <= count; i++) {
    try {
      const detectedAt = randomPastDate(60);
      const severity = randomElement(severities);
      const status = randomElement(statuses);
      const shouldResolve = status === 'resolved' || status === 'closed';

      // ステータスに応じた日時調整
      const detectedDate = new Date(detectedAt);
      const resolvedAt = shouldResolve
        ? new Date(detectedDate.getTime() + randomInt(1, 48) * 60 * 60 * 1000).toISOString()
        : null;

      const incident = {
        title: `${randomElement(titlePrefixes)} - ケース${i}`,
        description: randomElement(descriptions),
        severity: severity,
        status: status,
        impact_scope: randomElement(impactScopes),
        detected_at: detectedAt,
        resolved_at: resolvedAt,
        tag_ids: Array.from(
          { length: randomInt(1, 4) },
          () => randomElement(tags).id
        ).filter((v, i, a) => a.indexOf(v) === i), // 重複削除
        assignee_id: Math.random() > 0.3 ? randomElement(users).id : null
      };

      const response = await fetch(`${API_URL}/incidents`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify(incident)
      });

      if (response.ok) {
        const created = await response.json();
        createdIncidents.push(created);
        console.log(`✓ Created incident ${i}/${count}: ${incident.title}`);
      } else {
        const error = await response.text();
        console.log(`✗ Failed to create incident ${i}: ${error}`);
      }

      // APIへの負荷を軽減するため少し待機
      await new Promise(resolve => setTimeout(resolve, 100));
    } catch (err) {
      console.error(`Error creating incident ${i}:`, err.message);
    }
  }

  return createdIncidents;
}

// コメント追加
async function addCommentsToIncidents(token, incidents, users) {
  console.log(`\n=== Adding comments to incidents ===`);
  const comments = [
    '原因を調査中です。',
    '一時的な回避策を適用しました。',
    '関連チームに連絡しました。',
    '修正パッチを作成中です。',
    '再発防止策を検討しています。',
    'ログを解析した結果、原因を特定しました。',
    'ベンダーからの回答待ちです。',
    '影響範囲が拡大しています。',
    '復旧作業を開始しました。',
    'テスト環境で再現を確認しました。',
    'デプロイを一時停止しています。',
    '監視を強化しました。',
    '同様の事例を過去に確認しました。',
    'ドキュメントを更新しました。',
    '追加調査が必要です。'
  ];

  let totalComments = 0;

  for (const incident of incidents.slice(0, 30)) { // 最初の30件にコメント追加
    const numComments = randomInt(1, 5);

    for (let i = 0; i < numComments; i++) {
      try {
        const response = await fetch(`${API_URL}/incidents/${incident.id}/comments`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`
          },
          body: JSON.stringify({
            comment: randomElement(comments)
          })
        });

        if (response.ok) {
          totalComments++;
        } else {
          const error = await response.json().catch(() => ({ error: 'Unknown error' }));
          console.error(`Failed to add comment to incident ${incident.id}: ${error.error || response.statusText}`);
        }
      } catch (err) {
        console.error(`Error adding comment to incident ${incident.id}:`, err.message);
      }
    }
  }

  console.log(`✓ Added ${totalComments} comments to incidents`);
}

// メイン処理
async function main() {
  try {
    console.log('=== Incidex Test Data Generator ===');
    console.log(`API URL: ${API_URL}`);

    // 管理者でログイン
    console.log('\n=== Logging in as admin ===');
    const token = await login('admin@example.com', 'password123');
    console.log('✓ Logged in successfully');

    // 既存ユーザー取得
    const existingUsersResponse = await fetch(`${API_URL}/users`, {
      headers: { 'Authorization': `Bearer ${token}` }
    });
    const existingUsers = await existingUsersResponse.json();
    console.log(`\nExisting users: ${existingUsers.length}`);

    // ユーザー作成
    const newUsers = await createUsers(token, existingUsers.length, 20);
    const allUsers = [...existingUsers, ...newUsers];

    // 既存タグ取得
    const existingTagsResponse = await fetch(`${API_URL}/tags`, {
      headers: { 'Authorization': `Bearer ${token}` }
    });
    const existingTags = await existingTagsResponse.json();
    console.log(`\nExisting tags: ${existingTags.length}`);

    // タグ作成（既存のタグをスキップ）
    const newTags = await createTags(token, existingTags);
    const allTags = existingTags.length > 0 ? existingTags : newTags;

    // テンプレート作成
    const templates = await createTemplates(token, allTags, 10);

    // インシデント作成
    const incidents = await createIncidents(token, allUsers, allTags, 50);

    // コメント追加
    await addCommentsToIncidents(token, incidents, allUsers);

    console.log('\n=== Summary ===');
    console.log(`Users created: ${newUsers.length}`);
    console.log(`Total users: ${allUsers.length}`);
    console.log(`Tags created: ${newTags.length}`);
    console.log(`Total tags: ${allTags.length}`);
    console.log(`Templates created: ${templates.length}`);
    console.log(`Incidents created: ${incidents.length}`);
    console.log('\n✓ Test data generation completed!');

  } catch (err) {
    console.error('Error:', err);
    process.exit(1);
  }
}

main();
