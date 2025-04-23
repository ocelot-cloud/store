import {AxiosResponse} from "axios";
import {doRequest} from "@/components/requests";
import {hubBaseUrl} from "@/components/config";

export async function doHubRequest(path: string, data: any): Promise<(AxiosResponse | null)> {
    return doRequest(hubBaseUrl + "/api" + path, data)
}

export const defaultAllowedSymbols = '[0-9a-z]';
export const passwordAllowedSymbols = '[a-zA-Z0-9!@#$%&_,.?]';
export const versionAllowedSymbols = '[0-9a-z.]';
export const defaultMinLength = 3;
export const defaultMaxLength = 20;
export const minLengthPassword = 8;
export const maxLengthPassword = 30;

export function generateInvalidInputMessage(fieldName: string, allowedSymbols: string, minLength: number, maxLength: number): string {
    return `Invalid ${fieldName}, allowed symbols are ${allowedSymbols} and the length must be between ${minLength} and ${maxLength}.`;
}

export function getDefaultValidationRegex(): RegExp {
    return new RegExp(`^${defaultAllowedSymbols}{${defaultMinLength},${defaultMaxLength}}$`)
}