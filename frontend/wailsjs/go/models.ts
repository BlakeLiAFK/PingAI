export namespace checker {
	
	export class CheckResult {
	    item: string;
	    status: string;
	    latency: number;
	    ttft: number;
	    message: string;
	    detail: string;
	    tokenIn: number;
	    tokenOut: number;
	
	    static createFrom(source: any = {}) {
	        return new CheckResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.item = source["item"];
	        this.status = source["status"];
	        this.latency = source["latency"];
	        this.ttft = source["ttft"];
	        this.message = source["message"];
	        this.detail = source["detail"];
	        this.tokenIn = source["tokenIn"];
	        this.tokenOut = source["tokenOut"];
	    }
	}
	export class FullCheckResult {
	    providerID: string;
	    providerName: string;
	    baseURL: string;
	    model: string;
	    protocol: string;
	    results: CheckResult[];
	    modelList: string[];
	    startTime: string;
	    endTime: string;
	    totalLatency: number;
	
	    static createFrom(source: any = {}) {
	        return new FullCheckResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.providerID = source["providerID"];
	        this.providerName = source["providerName"];
	        this.baseURL = source["baseURL"];
	        this.model = source["model"];
	        this.protocol = source["protocol"];
	        this.results = this.convertValues(source["results"], CheckResult);
	        this.modelList = source["modelList"];
	        this.startTime = source["startTime"];
	        this.endTime = source["endTime"];
	        this.totalLatency = source["totalLatency"];
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
	
	export class AddProviderReq {
	    id: string;
	    name: string;
	    baseURL: string;
	    protocol: string;
	    models: string[];
	
	    static createFrom(source: any = {}) {
	        return new AddProviderReq(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.baseURL = source["baseURL"];
	        this.protocol = source["protocol"];
	        this.models = source["models"];
	    }
	}
	export class BatchCheckItem {
	    baseURL: string;
	    apiKey: string;
	    model: string;
	    providerID: string;
	    providerName: string;
	    protocol: string;
	
	    static createFrom(source: any = {}) {
	        return new BatchCheckItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.baseURL = source["baseURL"];
	        this.apiKey = source["apiKey"];
	        this.model = source["model"];
	        this.providerID = source["providerID"];
	        this.providerName = source["providerName"];
	        this.protocol = source["protocol"];
	    }
	}
	export class HistoryItem {
	    id: number;
	    providerID: string;
	    providerName: string;
	    baseURL: string;
	    model: string;
	    protocol: string;
	    results: checker.CheckResult[];
	    modelList: string[];
	    totalLatency: number;
	    status: string;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new HistoryItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.providerID = source["providerID"];
	        this.providerName = source["providerName"];
	        this.baseURL = source["baseURL"];
	        this.model = source["model"];
	        this.protocol = source["protocol"];
	        this.results = this.convertValues(source["results"], checker.CheckResult);
	        this.modelList = source["modelList"];
	        this.totalLatency = source["totalLatency"];
	        this.status = source["status"];
	        this.createdAt = source["createdAt"];
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
	export class HistoryListResult {
	    items: HistoryItem[];
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new HistoryListResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.items = this.convertValues(source["items"], HistoryItem);
	        this.total = source["total"];
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
	export class ProviderInfo {
	    id: string;
	    name: string;
	    baseURL: string;
	    protocol: string;
	    models: string[];
	    isBuiltin: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ProviderInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.baseURL = source["baseURL"];
	        this.protocol = source["protocol"];
	        this.models = source["models"];
	        this.isBuiltin = source["isBuiltin"];
	    }
	}

}

export namespace store {
	
	export class ProviderConfigRow {
	    providerID: string;
	    apiKey: string;
	    baseURL: string;
	    model: string;
	    protocol: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new ProviderConfigRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.providerID = source["providerID"];
	        this.apiKey = source["apiKey"];
	        this.baseURL = source["baseURL"];
	        this.model = source["model"];
	        this.protocol = source["protocol"];
	        this.updatedAt = source["updatedAt"];
	    }
	}

}

