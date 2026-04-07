export namespace analyse {
	
	export class SankeyLink {
	    source: string;
	    target: string;
	    value: number;
	
	    static createFrom(source: any = {}) {
	        return new SankeyLink(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.source = source["source"];
	        this.target = source["target"];
	        this.value = source["value"];
	    }
	}
	export class ConvoStats {
	    Total: number;
	    BySize: Record<string, number>;
	    Sankey: SankeyLink[];
	    GapMinutes: number;
	
	    static createFrom(source: any = {}) {
	        return new ConvoStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Total = source["Total"];
	        this.BySize = source["BySize"];
	        this.Sankey = this.convertValues(source["Sankey"], SankeyLink);
	        this.GapMinutes = source["GapMinutes"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DayCount {
	    date: string;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new DayCount(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.date = source["date"];
	        this.count = source["count"];
	    }
	}
	export class DomainCount {
	    domain: string;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new DomainCount(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.domain = source["domain"];
	        this.count = source["count"];
	    }
	}
	export class EmojiCount {
	    Emoji: string;
	    Count: number;
	
	    static createFrom(source: any = {}) {
	        return new EmojiCount(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Emoji = source["Emoji"];
	        this.Count = source["Count"];
	    }
	}
	export class GrowthPoint {
	    month: string;
	    counts: Record<string, number>;
	
	    static createFrom(source: any = {}) {
	        return new GrowthPoint(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.month = source["month"];
	        this.counts = source["counts"];
	    }
	}
	
	export class SentimentPoint {
	    month: string;
	    pos: number;
	    neg: number;
	    net: number;
	    netByAuthor: Record<string, number>;
	
	    static createFrom(source: any = {}) {
	        return new SentimentPoint(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.month = source["month"];
	        this.pos = source["pos"];
	        this.neg = source["neg"];
	        this.net = source["net"];
	        this.netByAuthor = source["netByAuthor"];
	    }
	}
	export class UserStats {
	    Messages: number;
	    Words: number;
	    UniqueWords: number;
	    Characters: number;
	    Emojis: number;
	    Laughs: number;
	    Apologies: number;
	    Questions: number;
	    Encouragement: number;
	    Images: number;
	    Videos: number;
	    Audios: number;
	    GIFs: number;
	    Stickers: number;
	    Links: number;
	    TopEmojis: EmojiCount[];
	    ConvosStarted: number;
	    ConvosClosed: number;
	    ConvosMissed: number;
	    TopContrib: number;
	    AvgConvoPts: number;
	    Reconnects: number;
	    DoubleMsgs: number;
	    RapidFirstPct: number;
	    AvgFirstResp: number;
	    AvgResponse: number;
	    MedianResp: number;
	    P90Response: number;
	    Positive: number;
	    Negative: number;
	    SentimentAvg: number;
	    CompoundAvg: number;
	    Emotion: number[];
	    IntensitySum: number[];
	    IntensityPeak: number;
	    VAD: number[];
	    VADMessages: number;
	    ScoredMsgs: number;
	    PeakHour: number;
	    Chronotype: string;
	    TopTerms: string[];
	    ReplyByHour: number[];
	
	    static createFrom(source: any = {}) {
	        return new UserStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Messages = source["Messages"];
	        this.Words = source["Words"];
	        this.UniqueWords = source["UniqueWords"];
	        this.Characters = source["Characters"];
	        this.Emojis = source["Emojis"];
	        this.Laughs = source["Laughs"];
	        this.Apologies = source["Apologies"];
	        this.Questions = source["Questions"];
	        this.Encouragement = source["Encouragement"];
	        this.Images = source["Images"];
	        this.Videos = source["Videos"];
	        this.Audios = source["Audios"];
	        this.GIFs = source["GIFs"];
	        this.Stickers = source["Stickers"];
	        this.Links = source["Links"];
	        this.TopEmojis = this.convertValues(source["TopEmojis"], EmojiCount);
	        this.ConvosStarted = source["ConvosStarted"];
	        this.ConvosClosed = source["ConvosClosed"];
	        this.ConvosMissed = source["ConvosMissed"];
	        this.TopContrib = source["TopContrib"];
	        this.AvgConvoPts = source["AvgConvoPts"];
	        this.Reconnects = source["Reconnects"];
	        this.DoubleMsgs = source["DoubleMsgs"];
	        this.RapidFirstPct = source["RapidFirstPct"];
	        this.AvgFirstResp = source["AvgFirstResp"];
	        this.AvgResponse = source["AvgResponse"];
	        this.MedianResp = source["MedianResp"];
	        this.P90Response = source["P90Response"];
	        this.Positive = source["Positive"];
	        this.Negative = source["Negative"];
	        this.SentimentAvg = source["SentimentAvg"];
	        this.CompoundAvg = source["CompoundAvg"];
	        this.Emotion = source["Emotion"];
	        this.IntensitySum = source["IntensitySum"];
	        this.IntensityPeak = source["IntensityPeak"];
	        this.VAD = source["VAD"];
	        this.VADMessages = source["VADMessages"];
	        this.ScoredMsgs = source["ScoredMsgs"];
	        this.PeakHour = source["PeakHour"];
	        this.Chronotype = source["Chronotype"];
	        this.TopTerms = source["TopTerms"];
	        this.ReplyByHour = source["ReplyByHour"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace main {
	
	export class PeriodDTO {
	    Start: string;
	    End: string;
	
	    static createFrom(source: any = {}) {
	        return new PeriodDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Start = source["Start"];
	        this.End = source["End"];
	    }
	}
	export class StatsDTO {
	    Participants: string[];
	    Period: PeriodDTO;
	    Messages: number;
	    Conversations: number;
	    ChatPoints: number;
	    PerUser: Record<string, analyse.UserStats>;
	    Timeline: analyse.GrowthPoint[];
	    Heatmap: number[][];
	    DailyActivity: analyse.DayCount[];
	    Convos: analyse.ConvoStats;
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
	    SentimentTimeline: analyse.SentimentPoint[];
	    TopDomains: analyse.DomainCount[];
	
	    static createFrom(source: any = {}) {
	        return new StatsDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Participants = source["Participants"];
	        this.Period = this.convertValues(source["Period"], PeriodDTO);
	        this.Messages = source["Messages"];
	        this.Conversations = source["Conversations"];
	        this.ChatPoints = source["ChatPoints"];
	        this.PerUser = this.convertValues(source["PerUser"], analyse.UserStats, true);
	        this.Timeline = this.convertValues(source["Timeline"], analyse.GrowthPoint);
	        this.Heatmap = source["Heatmap"];
	        this.DailyActivity = this.convertValues(source["DailyActivity"], analyse.DayCount);
	        this.Convos = this.convertValues(source["Convos"], analyse.ConvoStats);
	        this.Insights = source["Insights"];
	        this.Rating = source["Rating"];
	        this.RatingLabel = source["RatingLabel"];
	        this.Balance = source["Balance"];
	        this.BalancePct = source["BalancePct"];
	        this.TopWeekdayHr = source["TopWeekdayHr"];
	        this.CharsTyped = source["CharsTyped"];
	        this.TimeTyping = source["TimeTyping"];
	        this.Direction = source["Direction"];
	        this.LongestStreak = source["LongestStreak"];
	        this.CurrentStreak = source["CurrentStreak"];
	        this.TopTerms = source["TopTerms"];
	        this.SentimentTimeline = this.convertValues(source["SentimentTimeline"], analyse.SentimentPoint);
	        this.TopDomains = this.convertValues(source["TopDomains"], analyse.DomainCount);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

