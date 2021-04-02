export class Sound {
  id!: string;
  name!: string;

  public getPlayUrl (): string {
    return `/play/sound/${this.id}/`
  }
}
