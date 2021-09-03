export class Sound {
  id!: string;
  name!: string;

  public getPlayUrl (): string {
    return `/sound/sounds/${this.id}/play/`
  }
}
