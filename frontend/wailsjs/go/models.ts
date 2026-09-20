export namespace backend {
	
	export class ProxyConfig {
	    enabled: boolean;
	    type: string;
	    host: string;
	    port: number;
	
	    static createFrom(source: any = {}) {
	        return new ProxyConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.type = source["type"];
	        this.host = source["host"];
	        this.port = source["port"];
	    }
	}
	export class DownloadRequest {
	    url: string;
	    name: string;
	    saveDir: string;
	    threads: number;
	    variantUri: string;
	    format: string;
	    proxy: ProxyConfig;
	
	    static createFrom(source: any = {}) {
	        return new DownloadRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.url = source["url"];
	        this.name = source["name"];
	        this.saveDir = source["saveDir"];
	        this.threads = source["threads"];
	        this.variantUri = source["variantUri"];
	        this.format = source["format"];
	        this.proxy = this.convertValues(source["proxy"], ProxyConfig);
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
	export class HistoryItem {
	    id: number;
	    name: string;
	    url: string;
	    size: number;
	    status: string;
	    finishedAt: string;
	    duration: string;
	    error: string;
	    file: string;
	
	    static createFrom(source: any = {}) {
	        return new HistoryItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.url = source["url"];
	        this.size = source["size"];
	        this.status = source["status"];
	        this.finishedAt = source["finishedAt"];
	        this.duration = source["duration"];
	        this.error = source["error"];
	        this.file = source["file"];
	    }
	}
	
	export class Settings {
	    threads: number;
	    concurrent: number;
	    saveDir: string;
	    notify: boolean;
	    format: string;
	    proxyEnabled: boolean;
	    proxyType: string;
	    proxyHost: string;
	    proxyPort: number;
	    updateMode: string;
	    currentVersion: string;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.threads = source["threads"];
	        this.concurrent = source["concurrent"];
	        this.saveDir = source["saveDir"];
	        this.notify = source["notify"];
	        this.format = source["format"];
	        this.proxyEnabled = source["proxyEnabled"];
	        this.proxyType = source["proxyType"];
	        this.proxyHost = source["proxyHost"];
	        this.proxyPort = source["proxyPort"];
	        this.updateMode = source["updateMode"];
	        this.currentVersion = source["currentVersion"];
	    }
	}
	export class Variant {
	    uri: string;
	    bandwidth: number;
	    resolution: string;
	    codecs: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new Variant(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.uri = source["uri"];
	        this.bandwidth = source["bandwidth"];
	        this.resolution = source["resolution"];
	        this.codecs = source["codecs"];
	        this.name = source["name"];
	    }
	}

}

