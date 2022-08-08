import config from "config";
import axios, {AxiosError} from "axios";

const check: string = config.get("check_for_broken_references");

const cache: HttpCache = {};

function setCache(key: string, code: number) {
    cache[key] = code;
}

function getCache(key: string): number {
    return cache[key];
}

function logResponse(link: string, code: number, isFromCache: boolean) {
    const term = isFromCache ? "via cache" : "";
    console.log("Got", code, term, "for", link);
}

async function checkReference(link: string): Promise<boolean> {
    const code = getCache(link);
    if (!isNaN(code)) {
        logResponse(link, code, true);
        return code == 200;
    }
    try {
        const res = await axios.head(link)
        logResponse(link, res.status, false);
        setCache(link, res.status);
        return res.status == 200;
    } catch (err) {
        if (err instanceof AxiosError && err.response) {
            setCache(link, err.response.status);
            logResponse(link, err.response.status, false);
        } else {
            // If we got here, it means that we failed to reach the server because of things that are IN OUR CONTROL
            // (e.g. timeout, socket reset)
            console.log(err);
            process.abort();
        }
        return false;
    }
}

async function removeIfBroken(references: string[]) {
    for (let i = references.length - 1; i >= 0; i--)
        if (!(await checkReference(references[i])))
            references.splice(i, 1)
}

export async function FixBrokenReferences(parsed_benchmarks: BenchmarkSchema[]): Promise<BenchmarkSchema[]> {
    if (!check)
        return parsed_benchmarks;
    await Promise.all(parsed_benchmarks.map(
        async bench => await Promise.all(bench.rules.map(
            async rule => await removeIfBroken(rule.references)))))
    return parsed_benchmarks
}