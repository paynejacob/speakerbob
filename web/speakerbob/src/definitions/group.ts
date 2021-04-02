export class Group {
  id!: string;
  name!: string;
  sounds!: string[];

  public getPlayUrl (): string {
    return `/play/group/${this.id}/`
  }
}
