import { Moment } from 'moment';
import moment from 'moment';
import API from '@/utils/api';

export interface Macro {
  id: string
  created_at: Moment
  name: string
  play_count: number
  NSFW: boolean
}

export interface MacroForm {
  name: string
  NSFW: boolean
}

export interface MacroFilter {}

export function toMacro(raw: any): Macro {
  return {
    id: raw.id,
    created_at: moment(raw.created_at),
    name: raw.name,
    play_count: raw.play_count,
    NSFW: raw.nsfw,
  };
}

export class MacroAPI extends API<Macro, MacroForm, MacroFilter> {
  protected toInstance(raw: any): Macro {
    return toMacro(raw);
  }
}
