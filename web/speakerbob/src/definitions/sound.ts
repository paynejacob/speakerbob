export class Sound {
  id!: string;
  name!: string;
  nsfw!: boolean;

  public getPlayUrl (): string {
    return `/play/sound/${this.id}/`
  }
}
