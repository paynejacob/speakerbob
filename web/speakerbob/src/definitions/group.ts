export class Group {
  id!: string;
  name!: string;
  nsfw!: boolean;
  sounds!: string[];

  public getPlayUrl (): string {
    return `/play/group/${this.id}/`
  }
}
