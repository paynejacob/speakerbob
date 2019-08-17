import { Moment } from 'moment';
import moment from 'moment';
import API from '@/utils/api';

export interface Sound {
  id: string
  created_at: Moment
  name: string
  duration: number
  NSFW: boolean
  play_count: number
}

export interface SoundForm {
  name: string
  NSFW: boolean
}

export interface SoundFilter {}

export function toSound(raw: any): Sound {
  return {
    id: raw.id,
    created_at: moment(raw.created_at),
    name: raw.name,
    duration: raw.play_count,
    NSFW: raw.nsfw,
    play_count: raw.play_count,
  };
}

export class SoundAPI extends API<Sound, SoundForm, SoundFilter> {
  protected toInstance(raw: any): Sound {
    return toSound(raw);
  }
}
