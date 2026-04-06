// Lightweight ambient types for analyse.Stats. Wails will overwrite this
// file with auto-generated bindings on `wails dev` / `wails build`; the
// shapes here mirror internal/analyse/analyse.go closely enough for the UI.

export namespace analyse {
  export interface EmojiCount { Emoji: string; Count: number }
  export interface GrowthPoint { month: string; counts: Record<string, number> }
  export interface SentimentPoint { month: string; pos: number; neg: number; net: number }
  export interface DomainCount { domain: string; count: number }
  export interface DayCount { date: string; count: number }
  export interface SankeyLink { source: string; target: string; value: number }
  export interface ConvoStats {
    Total: number;
    BySize: Record<string, number>;
    Sankey: SankeyLink[];
    GapMinutes: number;
  }
  export interface UserStats {
    Messages: number; Words: number; UniqueWords: number; Characters: number;
    Emojis: number; Laughs: number; Apologies: number; Questions: number; Encouragement: number;
    Images: number; Videos: number; Audios: number; GIFs: number; Stickers: number; Links: number;
    TopEmojis: EmojiCount[];
    ConvosStarted: number; ConvosClosed: number; ConvosMissed: number;
    TopContrib: number; AvgConvoPts: number;
    Reconnects: number; DoubleMsgs: number; RapidFirstPct: number;
    AvgFirstResp: number; AvgResponse: number; MedianResp: number; P90Response: number;
    Positive: number; Negative: number; SentimentAvg: number;
    PeakHour: number; Chronotype: string;
    TopTerms: string[];
    ReplyByHour: number[];
  }
  export interface Stats {
    Participants: [string, string];
    Period: { Start: string; End: string };
    Messages: number;
    Conversations: number;
    ChatPoints: number;
    PerUser: Record<string, UserStats>;
    Timeline: GrowthPoint[];
    Heatmap: number[][];
    DailyActivity: DayCount[];
    Convos: ConvoStats;
    Insights: string[];
    Rating: number;
    RatingLabel: string;
    Balance: Record<string, number>;
    BalancePct: Record<string, number>;
    TopWeekdayHr: string;
    CharsTyped: number;
    TimeTyping: string;
    Direction: Record<string, number>;
    LongestStreak: number;
    CurrentStreak: number;
    TopTerms: string[];
    SentimentTimeline: SentimentPoint[];
    TopDomains: DomainCount[];
  }
}
