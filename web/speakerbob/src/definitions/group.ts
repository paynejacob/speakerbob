export class Group {
  id!: string;
  name!: string;
  sounds!: string[];

  public getPlayUrl (): string {
    return `/sound/groups/${this.id}/play/`
  }
}
