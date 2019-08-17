import { AxiosInstance, AxiosResponse } from 'axios';
import { joinPath } from '@/utils/helpers';

export interface ListResult<Model> {
  page: number;
  hasNext: boolean;
  hasPrevious: boolean;
  results: Model[];
}

export class UnexpectedStatusError extends Error {
  public readonly status: number;

  constructor(status: number) {
    super('');
    this.status = status;
  }
}

export class ValidationError extends Error {
  public errors: Map<string, string[]>;

  constructor(props: object) {
    super('');
    this.errors = new Map(Object.entries(props));
  }

  public getErrors(key: string): string[] {
    return this.errors.get(key) || [];
  }

  public hasErrors(key: string): boolean {
    return !!this.getErrors(key).length;
  }
}

export default abstract class API<Model, Form, Filter> {
  protected readonly service: AxiosInstance;

  protected readonly path: string;

  protected constructor(service: AxiosInstance, path: string) {
    this.service = service;
    this.path = path;
  }

  public async create(data: Form): Promise<Model> {
    const response: AxiosResponse = await this.service.post(this.getPath(), data);

    if (response.status === 400) {
      throw new ValidationError(response.data);
    } else if (response.status === 201) {
      return this.toInstance(response.data);
    } else {
      throw new UnexpectedStatusError(response.status);
    }
  }

  public async detail(pk: string): Promise<Model> {
    const response: AxiosResponse = await this.service.get(this.getPath(pk));

    if (response.status !== 200) {
      throw new UnexpectedStatusError(response.status);
    }

    return this.toInstance(response.data);
  }

  public async update(pk: string, data: Form): Promise<Model> {
    const response: AxiosResponse = await this.service.patch(this.getPath(pk), data);

    if (response.status === 400) {
      throw new ValidationError(response.data);
    } else if (response.status === 200) {
      return this.toInstance(response.data);
    } else {
      throw new UnexpectedStatusError(response.status);
    }
  }

  public async delete(pk: string): Promise<boolean> {
    const response = await this.service.delete(this.getPath(pk));

    if (response.status > 404) {
      throw new UnexpectedStatusError(response.status);
    }

    return response.status === 204;
  }

  public async list(page: number = 1, filters: Filter, sort: string[]): Promise<ListResult<Model>> {
    const response: AxiosResponse = await this.service.get(this.getPath(), {
      params: {
        page,
        ...filters,
        sort: sort.join(','),
      },
    });

    if (response.status !== 200) {
      throw new UnexpectedStatusError(response.status);
    }

    return {
      page,
      hasNext: response.data.next !== null,
      hasPrevious: response.data.previous !== null,
      results: response.data.results.map(this.toInstance),
    };
  }

  public async listAll(filter: Filter, sort: string[]): Promise<Model[]> {
    let page = 1;
    let hasNext: boolean = true;
    let results: Model[] = [];
    while (hasNext) {
      const listResult = await this.list(page, filter, sort);
      hasNext = listResult.hasNext;
      results = [...results, ...listResult.results];
      page++;
    }

    return results;
  }

  protected abstract toInstance(raw: any): Model;

  protected getPath(...parts: string[]): string {
    return joinPath(this.path, ...parts);
  }
}
